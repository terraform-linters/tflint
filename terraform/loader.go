package terraform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
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
}

// NewLoader creates and returns a loader that reads configuration from the
// given filesystem.
//
// The loader has some internal state about the modules that are currently
// installed, which is read from disk as part of this function. Note that
// this will always read against the current directory unless TF_DATA_DIR
// is set.
func NewLoader(fs afero.Afero) (*Loader, error) {
	log.Print("[INFO] Initialize new loader")

	ret := &Loader{
		parser: NewParser(fs),
		modules: moduleMgr{
			fs:       fs,
			manifest: moduleManifest{},
		},
	}

	err := ret.modules.readModuleManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to read module manifest: %s", err)
	}

	return ret, nil
}

// LoadConfig reads the Terraform module in the given directory and uses it as the
// root module to build the static module tree that represents a configuration.
//
// The second argument determines whether to load child modules. If true is given,
// load installed child modules according to a manifest file. If false is given,
// all child modules will not be loaded.
func (l *Loader) LoadConfig(dir string, module bool) (*Config, hcl.Diagnostics) {
	mod, diags := l.parser.LoadConfigDir(dir)
	if diags.HasErrors() {
		return nil, diags
	}

	var walker ModuleWalkerFunc
	if module {
		log.Print("[INFO] Module inspection is enabled. Building the root module with children...")
		walker = ModuleWalkerFunc(l.moduleWalkerLoad)
	} else {
		log.Print("[INFO] Module inspection is disabled. Building the root module without children...")
		walker = ModuleWalkerFunc(l.moduleWalkerIgnore)
	}

	cfg, diags := BuildConfig(mod, walker)
	if diags.HasErrors() {
		return nil, diags
	}
	return cfg, nil
}

func (l *Loader) moduleWalkerLoad(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) {
	// Since we're just loading here, we expect that all referenced modules
	// will be already installed and described in our manifest. However, we
	// do verify that the manifest and the configuration are in agreement
	// so that we can prompt the user to run "terraform init" if not.

	key := l.modules.manifest.moduleKey(req.Path)
	record, exists := l.modules.manifest[key]

	if !exists {
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
}

func (l *Loader) moduleWalkerIgnore(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) {
	// Prevents loading any child modules by returning nil for all module requests
	return nil, nil, nil
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

	autoLoadFiles, listDiags := l.parser.autoLoadValuesDirFiles(dir)
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
	vals, diags := l.parser.LoadValuesFile(file)
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
	return l.parser.LoadConfigDirFiles(dir)
}

func (l *Loader) Sources() map[string][]byte {
	return l.parser.Sources()
}

func (l *Loader) Files() map[string]*hcl.File {
	return l.parser.Files()
}
