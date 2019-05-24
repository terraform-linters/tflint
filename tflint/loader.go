package tflint

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
)

//go:generate mockgen -source loader.go -destination ../mock/loader.go -package mock

// AbstractLoader is a loader interface for mock
type AbstractLoader interface {
	LoadConfig() (*configs.Config, error)
	LoadValuesFiles(...string) ([]terraform.InputValues, error)
	IsConfigFile(string) bool
}

// Loader is a wrapper of Terraform's configload.Loader
type Loader struct {
	loader               *configload.Loader
	configFiles          []string
	moduleSourceVersions map[string][]*version.Version
	moduleManifest       map[string]*moduleManifest
}

type moduleManifest struct {
	Key        string           `json:"Key"`
	Source     string           `json:"Source"`
	Version    *version.Version `json:"-"`
	VersionStr string           `json:"Version,omitempty"`
	Dir        string           `json:"Dir"`
	Root       string           `json:"Root"`
}

type moduleManifestFile struct {
	Modules []*moduleManifest `json:"Modules"`
}

// NewLoader returns a loader with module manifests
func NewLoader() (*Loader, error) {
	log.Print("[INFO] Initialize new loader")

	loader, err := configload.NewLoader(&configload.Config{
		ModulesDir: getTFModuleDir(),
	})
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}

	primary, override, diags := loader.Parser().ConfigDirFiles(".")
	if diags != nil {
		log.Printf("[ERROR] %s", diags)
		return nil, diags
	}

	l := &Loader{
		loader:               loader,
		configFiles:          append(primary, override...),
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
func (l *Loader) LoadConfig() (*configs.Config, error) {
	log.Print("[INFO] Load configurations under the current directory")
	rootMod, diags := l.loader.Parser().LoadConfigDir(".")
	if diags.HasErrors() {
		log.Printf("[ERROR] %s", diags)
		return nil, diags
	}

	log.Print("[DEBUG] Trying to load modules using the legacy module walker...")
	cfg, diags := configs.BuildConfig(rootMod, l.moduleWalkerLegacy())
	if !diags.HasErrors() {
		return cfg, nil
	}
	log.Print("[DEBUG] Failed to load modules using the legacy module walker; Trying the v0.10.6 module walker...")
	log.Printf("[DEBUG] Original error: %s", diags)

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_10_6())
	if !diags.HasErrors() {
		return cfg, nil
	}
	log.Print("[DEBUG] Failed to load modules using the v0.10.6 module walker; Trying the v0.10.7 ~ v0.10.8 module walker...")
	log.Printf("[DEBUG] Original error: %s", diags)

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_10_7V0_10_8())
	if !diags.HasErrors() {
		return cfg, nil
	}
	log.Print("[DEBUG] Failed to load modules using the v0.10.7 ~ v0.10.8 module walker; Trying the v0.11.0 ~ v0.11.7 module walker...")
	log.Printf("[DEBUG] Original error: %s", diags)

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_11_0V0_11_7())
	if !diags.HasErrors() {
		return cfg, nil
	}
	log.Printf("[ERROR] Failed to load modules using the v0.11.0 ~ v0.11.7 module walker; Trying the v0.12 module walker...")
	log.Printf("[DEBUG] Original error: %s", diags)

	cfg, diags = configs.BuildConfig(rootMod, l.moduleWalkerV0_12())
	if !diags.HasErrors() {
		return cfg, nil
	}

	log.Printf("[ERROR] Failed to load modules using the v0.12 module walker: %s", diags)
	return nil, diags
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

	autoLoadFiles, err := autoLoadValuesFiles()
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

// IsConfigFile checks whether the configuration files includes the file
func (l *Loader) IsConfigFile(file string) bool {
	for _, configFile := range l.configFiles {
		if file == configFile {
			return true
		}
	}
	return false
}

// autoLoadValuesFiles returns all files which match *.auto.tfvars present in the current directory
// The list is sorted alphabetically. This is equivalent to priority
// Please note that terraform.tfvars is not included in this list
func autoLoadValuesFiles() ([]string, error) {
	files, err := ioutil.ReadDir(".")
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
	vals, diags := l.loader.Parser().LoadValuesFile(file)
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

func (l *Loader) moduleWalkerLegacy() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		key := "root." + req.Name + "-" + req.SourceAddr
		dir := makeModuleDirFromKey(key)
		log.Printf("[DEBUG] Trying to load the module: key=%s, dir=%s", key, dir)
		mod, diags := l.loader.Parser().LoadConfigDir(dir)
		return mod, nil, diags
	})
}

func (l *Loader) moduleWalkerV0_10_6() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		key := "module." + req.Name + "-" + req.SourceAddr
		dir := makeModuleDirFromKey(key)
		log.Printf("[DEBUG] Trying to load the module: key=%s, dir=%s", key, dir)
		mod, diags := l.loader.Parser().LoadConfigDir(dir)
		return mod, nil, diags
	})
}

func (l *Loader) moduleWalkerV0_10_7V0_10_8() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		key := "0.root." + req.Name + "-" + req.SourceAddr
		dir := makeModuleDirFromKey(key)
		log.Printf("[DEBUG] Trying to load the module: key=%s, dir=%s", key, dir)
		mod, diags := l.loader.Parser().LoadConfigDir(dir)
		return mod, nil, diags
	})
}

func (l *Loader) moduleWalkerV0_11_0V0_11_7() configs.ModuleWalker {
	return configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
		path := append(buildParentModulePathTree([]string{}, req.Parent), l.getModulePath(req))
		key := "1." + strings.Join(path, "|")

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

		dir := record.Dir
		if record.Root != "" {
			dir = filepath.Join(dir, record.Root)
		}
		log.Printf("[DEBUG] Trying to load the module: key=%s, version=%s, dir=%s", key, record.VersionStr, dir)

		mod, diags := l.loader.Parser().LoadConfigDir(dir)
		return mod, record.Version, diags
	})
}

func (l *Loader) moduleWalkerV0_12() configs.ModuleWalker {
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

		dir := record.Dir
		if record.Root != "" {
			dir = filepath.Join(dir, record.Root)
		}
		log.Printf("[DEBUG] Trying to load the module: key=%s, version=%s, dir=%s", key, record.VersionStr, dir)

		mod, diags := l.loader.Parser().LoadConfigDir(dir)
		return mod, record.Version, diags
	})
}

func (l *Loader) initializeModuleManifest() error {
	file, err := ioutil.ReadFile(getTFModuleManifestPath())
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
		l.moduleManifest[m.Key] = m
	}

	return nil
}

func makeModuleDirFromKey(key string) string {
	sum := md5.Sum([]byte(key))
	return filepath.Join(getTFModuleDir(), hex.EncodeToString(sum[:]))
}

func buildParentModulePathTree(path []string, cfg *configs.Config) []string {
	if cfg.Path.IsRoot() {
		// @see https://github.com/golang/go/wiki/SliceTricks#reversing
		for i := len(path)/2 - 1; i >= 0; i-- {
			opp := len(path) - 1 - i
			path[i], path[opp] = path[opp], path[i]
		}
		return path
	}

	_, call := cfg.Path.Call()
	key := call.Name
	if cfg.Version != nil {
		key += "#" + cfg.Version.String()
	}
	key += ";" + cfg.SourceAddr
	path = append(path, key)

	return buildParentModulePathTree(path, cfg.Parent)
}

func (l *Loader) getModulePath(req *configs.ModuleRequest) string {
	key := req.Name + ";" + req.SourceAddr
	if len(req.VersionConstraint.Required) > 0 {
		log.Printf("[DEBUG] Processing the `%s` module: constraints=%#v", req.Name, req.VersionConstraint)
		sourceVersions := l.moduleSourceVersions[req.SourceAddr]

		var latest *version.Version
		for _, v := range sourceVersions {
			if req.VersionConstraint.Required.Check(v) {
				if latest == nil || v.GreaterThan(latest) {
					latest = v
				}
			} else {
				log.Printf("[INFO] `%s` doesn't satisfy the version constraint. Ignored.", v)
			}
		}

		if latest == nil {
			panic(fmt.Errorf("There is no version that satisfies the constraints: name=%s, constraints=%#v, versions=%#v", req.Name, req.VersionConstraint, l.moduleSourceVersions[req.SourceAddr]))
		}
		key += "." + latest.String()
	}

	return key
}
