package loader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	hclParser "github.com/hashicorp/hcl/hcl/parser"
	jsonParser "github.com/hashicorp/hcl/json/parser"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
)

type LoaderIF interface {
	LoadTemplate(filename string) error
	LoadModuleFile(moduleKey string, source string) error
	LoadAllTemplate(dir string) error
	Dump() (map[string]*ast.File, *state.TFState, []*ast.File)
	LoadState()
	LoadTFVars([]string)
}

type Loader struct {
	Logger    *logger.Logger
	Templates map[string]*ast.File
	State     *state.TFState
	TFVars    []*ast.File
}

func NewLoader(debug bool) *Loader {
	return &Loader{
		Logger:    logger.Init(debug),
		Templates: make(map[string]*ast.File),
		State:     &state.TFState{},
		TFVars:    []*ast.File{},
	}
}

func (l *Loader) LoadTemplate(filename string) error {
	root, err := loadHCL(filename, l.Logger)
	if err != nil {
		return err
	}

	l.Templates[filename] = root
	return nil
}

func (l *Loader) LoadModuleFile(moduleKey string, source string) error {
	l.Logger.Info(fmt.Sprintf("load module `%s`", source))
	modulePath := ".terraform/modules/" + moduleKey
	if _, err := os.Stat(modulePath); err != nil {
		l.Logger.Error(err)
		return fmt.Errorf("ERROR: module `%s` not found. Did you run `terraform get`?", source)
	}
	filePattern := modulePath + "/*.tf"
	files, err := filepath.Glob(filePattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		root, err := loadHCL(file, l.Logger)
		if err != nil {
			return err
		}
		filename := strings.Replace(file, modulePath, "", 1)
		fileKey := source + filename
		l.Templates[fileKey] = root
	}

	return nil
}

func (l *Loader) LoadAllTemplate(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return err
	}
	filePattern := dir + "/*.tf"
	files, err := filepath.Glob(filePattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := l.LoadTemplate(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Loader) LoadState() {
	l.Logger.Info("Load tfstate...")
	var statePath string
	// stat local state
	if _, err := os.Stat(state.LocalStatePath); err != nil {
		l.Logger.Error(err)
		// stat remote state
		if _, err := os.Stat(state.RemoteStatePath); err != nil {
			l.Logger.Error(err)
			return
		} else {
			l.Logger.Info("Remote state detected")
			statePath = state.RemoteStatePath
		}
	} else {
		l.Logger.Info("Local state detected")
		statePath = state.LocalStatePath
	}

	jsonBytes, err := ioutil.ReadFile(statePath)
	if err != nil {
		l.Logger.Error(err)
		return
	}
	if err := json.Unmarshal(jsonBytes, l.State); err != nil {
		l.Logger.Error(err)
		return
	}
}

func (l *Loader) LoadTFVars(varfile []string) {
	l.Logger.Info("Load tfvars...")

	for _, file := range varfile {
		l.Logger.Info(fmt.Sprintf("Load `%s`", file))
		if _, err := os.Stat(file); err != nil {
			l.Logger.Error(err)
			continue
		}

		var tfvar *ast.File
		var err error
		tfvar, err = loadHCL(file, l.Logger)
		if err != nil {
			l.Logger.Error(err)
			tfvar, err = loadJSON(file, l.Logger)
			if err != nil {
				l.Logger.Error(err)
				continue
			}
		}

		l.TFVars = append(l.TFVars, tfvar)
	}
}

func (l *Loader) Dump() (map[string]*ast.File, *state.TFState, []*ast.File) {
	return l.Templates, l.State, l.TFVars
}

func loadHCL(filename string, l *logger.Logger) (*ast.File, error) {
	l.Info(fmt.Sprintf("Load HCL file: `%s`", filename))
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("ERROR: Cannot open file %s", filename)
	}
	root, err := hclParser.Parse(b)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("ERROR: Parse error occurred by %s", filename)
	}

	return root, nil
}

func loadJSON(filename string, l *logger.Logger) (*ast.File, error) {
	l.Info(fmt.Sprintf("load JSON file: `%s`", filename))
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("ERROR: Cannot open file %s", filename)
	}
	root, err := jsonParser.Parse(b)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("ERROR: Parse error occurred by %s", filename)
	}

	return root, nil
}
