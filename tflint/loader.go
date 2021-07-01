package tflint

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/terraform"
)

//go:generate go run github.com/golang/mock/mockgen -source loader.go -destination loader_mock.go -package tflint -self_package github.com/terraform-linters/tflint/tflint

// AbstractLoader is a loader interface for mock
type AbstractLoader interface {
	LoadConfig(string) (*configs.Config, error)
	LoadAnnotations(string) (map[string]Annotations, error)
	LoadValuesFiles(...string) ([]terraform.InputValues, error)
	Files() (map[string]*hcl.File, error)
	Sources() map[string][]byte
}

// Loader is a wrapper of Terraform's configload.Loader
type Loader struct {
	parser               *configs.Parser
	fs                   afero.Afero
	currentDir           string
	config               *Config
	moduleSourceVersions map[string][]*version.Version
	moduleManifest       map[string]*moduleManifest
}

type moduleManifest struct {
	Key        string           `json:"Key"`
	Source     string           `json:"Source"`
	Version    *version.Version `json:"-"`
	VersionStr string           `json:"Version,omitempty"`
	Dir        string           `json:"Dir"`
}

type moduleManifestFile struct {
	Modules []*moduleManifest `json:"Modules"`
}

// NewLoader returns a loader with module manifests
func NewLoader(fs afero.Afero, cfg *Config) (*Loader, error) {
	log.Print("[INFO] Initialize new loader")

	l := &Loader{
		parser:               configs.NewParser(fs),
		fs:                   fs,
		config:               cfg,
		moduleSourceVersions: map[string][]*version.Version{},
		moduleManifest:       map[string]*moduleManifest{},
	}

	if _, err := os.Stat(getTFModuleManifestPath()); !os.IsNotExist(err) {
		log.Print("[INFO] Module manifest file found. Initializing...")
		if err := l.initializeModuleManifest(); err != nil {
			log.Printf("[ERROR] %s", err)
			return nil, err
		}
	}

	return l, nil
}

// LoadConfig loads Terraform's configurations
// TODO: Can we use configload.LoadConfig instead?
func (l *Loader) LoadConfig(dir string) (*configs.Config, error) {
	l.currentDir = dir
	log.Printf("[INFO] Load configurations under %s", dir)
	rootMod, diags := l.parser.LoadConfigDir(dir)
	if diags.HasErrors() {
		log.Printf("[ERROR] %s", diags)
		return nil, diags
	}

	if !l.config.Module {
		log.Print("[INFO] Module inspection is disabled. Building a root module without children...")
		cfg, diags := configs.BuildConfig(rootMod, l.ignoreModuleWalker())
		if diags.HasErrors() {
			return nil, diags
		}
		return cfg, nil
	}
	log.Print("[INFO] Module inspection is enabled. Building a root module with children...")

	cfg, diags := configs.BuildConfig(rootMod, l.moduleWalker())
	if !diags.HasErrors() {
		return cfg, nil
	}

	log.Printf("[ERROR] Failed to load modules: %s", diags)
	return nil, diags
}

// Files returns a map of hcl.File pointers for every file that has been read by the loader.
// It uses the source cache to avoid re-loading the files from disk. These files can be used
// to do low level decoding of Terraform configuration.
func (l *Loader) Files() (map[string]*hcl.File, error) {
	sources := l.parser.Sources()
	result := make(map[string]*hcl.File, len(sources))
	parser := hclparse.NewParser()

	for path, src := range sources {
		var file *hcl.File
		var diags hcl.Diagnostics
		switch {
		case strings.HasSuffix(path, ".json"):
			file, diags = parser.ParseJSON(src, path)
		default:
			file, diags = parser.ParseHCL(src, path)
		}

		if diags.HasErrors() {
			return nil, diags
		}

		result[path] = file
	}

	return result, nil
}

// LoadAnnotations load TFLint annotation comments as HCL tokens.
func (l *Loader) LoadAnnotations(dir string) (map[string]Annotations, error) {
	primary, override, diags := l.parser.ConfigDirFiles(dir)
	if diags != nil {
		log.Printf("[ERROR] %s", diags)
		return nil, diags
	}
	configFiles := append(primary, override...)

	ret := map[string]Annotations{}

	for _, configFile := range configFiles {
		if !strings.HasSuffix(configFile, ".tf") {
			continue
		}

		src, err := l.fs.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		tokens, diags := hclsyntax.LexConfig(src, configFile, hcl.Pos{Byte: 0, Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, diags
		}
		ret[configFile] = NewAnnotations(tokens)
	}

	return ret, nil
}

// LoadValuesFiles reads Terraform's values files and returns terraform.InputValues list in order of priority
// Pass values ​​files specified from the CLI as the arguments in order of priority
// This is the responsibility of the caller
func (l *Loader) LoadValuesFiles(files ...string) ([]terraform.InputValues, error) {
	log.Print("[INFO] Load values files")

	values := []terraform.InputValues{}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return values, fmt.Errorf("`%s` is not found", file)
		}
	}

	autoLoadFiles, err := l.autoLoadValuesFiles()
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	if _, err := os.Stat(defaultValuesFile); !os.IsNotExist(err) {
		autoLoadFiles = append([]string{defaultValuesFile}, autoLoadFiles...)
	}

	for _, file := range autoLoadFiles {
		vals, err := l.loadValuesFile(file, terraform.ValueFromAutoFile)
		if err != nil {
			return nil, err
		}
		values = append(values, vals)
	}
	for _, file := range files {
		vals, err := l.loadValuesFile(file, terraform.ValueFromNamedFile)
		if err != nil {
			return nil, err
		}
		values = append(values, vals)
	}

	return values, nil
}

// Sources returns the source code cache for the underlying parser of this loader
func (l *Loader) Sources() map[string][]byte {
	return l.parser.Sources()
}

// autoLoadValuesFiles returns all files which match *.auto.tfvars present in the current directory
// The list is sorted alphabetically. This is equivalent to priority
// Please note that terraform.tfvars is not included in this list
func (l *Loader) autoLoadValuesFiles() ([]string, error) {
	files, err := l.fs.ReadDir(".")
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".auto.tfvars") || strings.HasSuffix(file.Name(), ".auto.tfvars.json") {
			ret = append(ret, file.Name())
		}
	}
	sort.Strings(ret)

	return ret, nil
}

func (l *Loader) loadValuesFile(file string, sourceType terraform.ValueSourceType) (terraform.InputValues, error) {
	log.Printf("[INFO] Load `%s`", file)
	vals, diags := l.parser.LoadValuesFile(file)
	if diags.HasErrors() {
		log.Printf("[ERROR] %s", diags)
		if diags[0].Subject == nil {
			// HACK: When Subject is nil, it outputs unintended message, so it replaces with actual file.
			return nil, errors.New(strings.Replace(diags.Error(), "<nil>: ", fmt.Sprintf("%s: ", file), 1))
		}
		return nil, diags
	}

	ret := make(terraform.InputValues)
	for k, v := range vals {
		ret[k] = &terraform.InputValue{
			Value:      v,
			SourceType: sourceType,
		}
	}
	return ret, nil
}

func (l *Loader) moduleWalker() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		key := req.Path.String()
		record, ok := l.moduleManifest[key]
		if !ok {
			log.Printf("[DEBUG] Failed to search by `%s` key.", key)
			return nil, nil, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("`%s` module is not found. Did you run `terraform init`?", req.Name),
					Subject:  &req.CallRange,
				},
			}
		}

		log.Printf("[DEBUG] Trying to load the module: key=%s, version=%s, dir=%s", key, record.VersionStr, record.Dir)

		mod, diags := l.parser.LoadConfigDir(record.Dir)
		return mod, record.Version, diags
	})
}

func (l *Loader) ignoreModuleWalker() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		return nil, nil, nil
	})
}

func (l *Loader) initializeModuleManifest() error {
	file, err := l.fs.ReadFile(getTFModuleManifestPath())
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Parsing the module manifest file: %s", file)

	var manifestFile moduleManifestFile
	err = json.Unmarshal(file, &manifestFile)
	if err != nil {
		return err
	}

	for _, m := range manifestFile.Modules {
		if m.VersionStr != "" {
			m.Version, err = version.NewVersion(m.VersionStr)
			if err != nil {
				return err
			}
			l.moduleSourceVersions[m.Source] = append(l.moduleSourceVersions[m.Source], m.Version)
		}

		moduleAddr := addrs.Module(strings.Split(m.Key, "."))
		l.moduleManifest[moduleAddr.String()] = m
	}

	return nil
}
