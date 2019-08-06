package tflint

import (
	"io/ioutil"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/terraform"
)

// Workspace represents the files being inspected and environments.
// It can be initialized and updated from the file system and external inputs,
// and is responsible for converting to Terraform built-in representations
// such as `configs.Config` and `terraform.InputValues`.
type Workspace struct {
	SourceDir   string
	ConfigFiles map[string]*ConfigFile
	ValuesFiles map[string]*ValuesFile
	Config      *Config

	loader *Loader
}

// ConfigFile is an intermidiate representation of a Terraform configuration.
type ConfigFile struct {
	Path     string
	Src      []byte
	Override bool
}

// ValuesFile is an intermidiate representation of a Terraform value file.
type ValuesFile struct {
	Path       string
	Src        []byte
	SourceType terraform.ValueSourceType
}

// LoadWorkspace loads the files under the given path and returns a new workspace.
func LoadWorkspace(config *Config, path string) (*Workspace, error) {
	// FIXME: Replace with the `configload.Loader`
	// FIXME: Set ModulesDir
	loader, err := NewLoader(config)
	if err != nil {
		return nil, err
	}

	ws := &Workspace{
		SourceDir:   path,
		ConfigFiles: map[string]*ConfigFile{},
		ValuesFiles: map[string]*ValuesFile{},
		Config:      config,

		loader: loader,
	}

	if err := ws.loadConfigFiles(); err != nil {
		return nil, err
	}
	if err := ws.loadValuesFiles(); err != nil {
		return nil, err
	}

	return ws, nil
}

func (w *Workspace) loadConfigFiles() error {
	primaryPaths, overridePaths, diags := w.loader.loader.Parser().ConfigDirFiles(w.SourceDir)
	if diags.HasErrors() {
		return diags
	}

	for _, path := range primaryPaths {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		w.ConfigFiles[path] = &ConfigFile{
			Path: path,
			Src:  src,
		}
	}

	for _, path := range overridePaths {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		w.ConfigFiles[path] = &ConfigFile{
			Path:     path,
			Src:      src,
			Override: true,
		}
	}

	return nil
}

func (w *Workspace) loadValuesFiles() error {
	autoLoadFiles, err := getAutoLoadValuesFiles(w.SourceDir)
	if err != nil {
		return err
	}

	for _, path := range autoLoadFiles {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		w.ValuesFiles[path] = &ValuesFile{
			Path:       path,
			Src:        src,
			SourceType: terraform.ValueFromAutoFile,
		}
	}

	for _, path := range w.Config.Varfile {
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		w.ValuesFiles[path] = &ValuesFile{
			Path:       path,
			Src:        src,
			SourceType: terraform.ValueFromNamedFile,
		}
	}

	return nil
}

// Update updates the workspace with the given file.
func (w *Workspace) Update(src []byte, path string) {
	if isConfigFile(path) {
		w.ConfigFiles[path] = &ConfigFile{
			Path:     path,
			Src:      src,
			Override: isOverrideConfigFile(path),
		}
	} else if isValuesFile(path) {
		if isAutoloadValuesFile(path) {
			w.ValuesFiles[path] = &ValuesFile{
				Path:       path,
				Src:        src,
				SourceType: terraform.ValueFromAutoFile,
			}
		}
		// TODO: Handle named values file
	}
}

// BuildConfig converts the workspace to `configs.Config`.
func (w *Workspace) BuildConfig() (*configs.Config, error) {
	var primaryFiles, overrideFiles []*configs.File
	for _, file := range w.ConfigFiles {
		body, diags := ParseHCLFile(file.Path, file.Src)
		if diags.HasErrors() || body == nil {
			return nil, diags
		}
		configFile, diags := BuildConfigFile(body, file.Override)
		if diags.HasErrors() {
			return nil, diags
		}

		if file.Override {
			overrideFiles = append(overrideFiles, configFile)
		} else {
			primaryFiles = append(primaryFiles, configFile)
		}
	}

	mod, diags := configs.NewModule(primaryFiles, overrideFiles)
	if diags.HasErrors() {
		return nil, diags
	}
	mod.SourceDir = w.SourceDir

	if !w.Config.Module {
		cfg, diags := configs.BuildConfig(mod, w.loader.ignoreModuleWalker())
		if diags.HasErrors() {
			return nil, diags
		}
		return cfg, nil
	}

	cfg, diags := configs.BuildConfig(mod, w.loader.moduleWalkerV0_12())
	if diags.HasErrors() {
		return nil, diags
	}

	return cfg, nil
}

// BuildAnnotations converts the workspace to `Annotations`
func (w *Workspace) BuildAnnotations() (map[string]Annotations, error) {
	ret := map[string]Annotations{}

	for _, file := range w.ConfigFiles {
		tokens, diags := hclsyntax.LexConfig(file.Src, file.Path, hcl.Pos{Byte: 0, Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, diags
		}
		ret[file.Path] = NewAnnotations(tokens)
	}

	return ret, nil
}

// BuildValuesFiles converts the workspace to `terraform.InputValues`
func (w *Workspace) BuildValuesFiles() ([]terraform.InputValues, error) {
	ret := []terraform.InputValues{}

	for _, file := range w.ValuesFiles {
		body, diags := ParseHCLFile(file.Path, file.Src)
		if diags.HasErrors() || body == nil {
			return nil, diags
		}
		valuesFile, diags := BuildValuesFile(body)
		if diags.HasErrors() {
			return nil, diags
		}

		inputValues := make(terraform.InputValues)
		for k, v := range valuesFile {
			inputValues[k] = &terraform.InputValue{
				Value:      v,
				SourceType: file.SourceType,
			}
		}
		ret = append(ret, inputValues)
	}

	return ret, nil
}
