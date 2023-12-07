package terraform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform/addrs"
)

// Loader is a fork of configload.Loader. The instance is the main entry-point
// for loading configurations via this package.
//
// It extends the general config-loading functionality in the Parser to support
// loading full configurations using modules and gathering input values from
// values files.
type Loader struct {
	parser  *Parser
	modules moduleMgr

	baseDir string
}

// NewLoader creates and returns a loader that reads configuration from the
// given filesystem.
//
// The loader has some internal state about the modules that are currently
// installed, which is read from disk as part of this function. Note that
// this will always read against the current directory unless TF_DATA_DIR
// is set.
//
// If an original working dir is passed, the paths of the loaded files will
// be relative to that directory.
func NewLoader(fs afero.Afero, originalWd string) (*Loader, error) {
	log.Print("[INFO] Initialize new loader")

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to determine current working directory: %s", err)
	}
	baseDir, err := filepath.Rel(originalWd, wd)
	if err != nil {
		return nil, fmt.Errorf("failed to determine base dir: %s", err)
	}

	ret := &Loader{
		parser: NewParser(fs),
		modules: moduleMgr{
			fs:       fs,
			manifest: moduleManifest{},
		},
		baseDir: baseDir,
	}

	err = ret.modules.readModuleManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read module manifest: %s", err)
	}

	return ret, nil
}

// LoadConfig reads the Terraform module in the given directory and uses it as the
// root module to build the static module tree that represents a configuration.
func (l *Loader) LoadConfig(dir string, callModuleType CallModuleType) (*Config, hcl.Diagnostics) {
	mod, diags := l.parser.LoadConfigDir(l.baseDir, dir)
	if diags.HasErrors() {
		return nil, diags
	}

	var walker ModuleWalkerFunc
	switch callModuleType {
	case CallAllModule:
		log.Print("[INFO] Building the root module while calling child modules...")
		walker = l.moduleWalkerFunc(true, true)
	case CallLocalModule:
		log.Print("[INFO] Building the root module while calling local child modules...")
		walker = l.moduleWalkerFunc(true, false)
	case CallNoModule:
		walker = l.moduleWalkerFunc(false, false)
	default:
		panic(fmt.Sprintf("unexpected module call type: %d", callModuleType))
	}

	cfg, diags := BuildConfig(mod, walker)
	if diags.HasErrors() {
		return nil, diags
	}
	return cfg, nil
}

func (l *Loader) moduleWalkerFunc(walkLocal, walkRemote bool) ModuleWalkerFunc {
	return func(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) {
		switch source := req.SourceAddr.(type) {
		case addrs.ModuleSourceLocal:
			if !walkLocal {
				return nil, nil, nil
			}
			dir := filepath.ToSlash(filepath.Join(req.Parent.Module.SourceDir, source.String()))
			log.Printf("[DEBUG] Trying to load the local module: name=%s dir=%s", req.Name, dir)
			if !l.parser.Exists(dir) {
				return nil, nil, hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf(`"%s" module is not found`, req.Name),
						Detail:   fmt.Sprintf(`The module directory "%s" does not exist or cannot be read.`, filepath.Join(l.baseDir, dir)),
						Subject:  &req.CallRange,
					},
				}
			}
			mod, diags := l.parser.LoadConfigDir(l.baseDir, dir)
			return mod, nil, diags

		case addrs.ModuleSourceRemote:
			if !walkRemote {
				return nil, nil, nil
			}
			// Since we're just loading here, we expect that all referenced modules
			// will be already installed and described in our manifest. However, we
			// do verify that the manifest and the configuration are in agreement
			// so that we can prompt the user to run "terraform init" if not.
			key := l.modules.manifest.moduleKey(req.Path)
			record, exists := l.modules.manifest[key]
			if !exists {
				log.Printf(`[DEBUG] Failed to find "%s"`, key)
				return nil, nil, hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf(`"%s" module is not found. Did you run "terraform init"?`, req.Name),
						Subject:  &req.CallRange,
					},
				}
			}
			log.Printf("[DEBUG] Trying to load the remote module: key=%s, version=%s, dir=%s", key, record.VersionStr, record.Dir)
			mod, diags := l.parser.LoadConfigDir(l.baseDir, record.Dir)
			return mod, record.Version, diags

		default:
			panic(fmt.Sprintf("unexpected module source type: %T", req.SourceAddr))
		}
	}
}

var defaultVarsFilename = "terraform.tfvars"

// LoadValuesFiles reads Terraform's autoloaded values files in the given directory
// and returns terraform.InputValues in order of priority.
//
// The second and subsequent arguments are given the paths of value files to be read
// manually. Argument order matches precedence.
func (l *Loader) LoadValuesFiles(dir string, files ...string) ([]InputValues, hcl.Diagnostics) {
	values := []InputValues{}
	diags := hcl.Diagnostics{}

	autoLoadFiles, listDiags := l.parser.autoLoadValuesDirFiles(l.baseDir, dir)
	diags = diags.Extend(listDiags)
	if listDiags.HasErrors() {
		return nil, diags
	}
	defaultVarsFile := filepath.Join(dir, defaultVarsFilename)
	if _, err := os.Stat(defaultVarsFile); err == nil {
		autoLoadFiles = append([]string{defaultVarsFile}, autoLoadFiles...)
	}

	for _, file := range autoLoadFiles {
		vals, loadDiags := l.loadValuesFile(file)
		diags = diags.Extend(loadDiags)
		if !loadDiags.HasErrors() {
			values = append(values, vals)
		}
	}
	for _, file := range files {
		vals, loadDiags := l.loadValuesFile(file)
		diags = diags.Extend(loadDiags)
		if !loadDiags.HasErrors() {
			values = append(values, vals)
		}
	}

	return values, diags
}

func (l *Loader) loadValuesFile(file string) (InputValues, hcl.Diagnostics) {
	vals, diags := l.parser.LoadValuesFile(l.baseDir, file)
	if diags.HasErrors() {
		return nil, diags
	}

	ret := make(InputValues)
	for k, v := range vals {
		ret[k] = &InputValue{
			Value: v,
		}
	}
	return ret, nil
}

func (l *Loader) LoadConfigDirFiles(dir string) (map[string]*hcl.File, hcl.Diagnostics) {
	return l.parser.LoadConfigDirFiles(l.baseDir, dir)
}

func (l *Loader) IsConfigDir(path string) bool {
	return l.parser.IsConfigDir(l.baseDir, path)
}

func (l *Loader) Sources() map[string][]byte {
	return l.parser.Sources()
}

func (l *Loader) Files() map[string]*hcl.File {
	return l.parser.Files()
}
