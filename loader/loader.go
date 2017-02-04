package loader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
)

type LoaderIF interface {
	LoadFile(filename string) error
	LoadModuleFile(moduleKey string, source string) error
	LoadAllFile(dir string) error
	DumpFiles() map[string]*ast.ObjectList
	LoadState()
}

type Loader struct {
	Logger  *logger.Logger
	ListMap map[string]*ast.ObjectList
	State   *state.TFState
}

func NewLoader(debug bool) *Loader {
	return &Loader{
		Logger:  logger.Init(debug),
		ListMap: make(map[string]*ast.ObjectList),
		State:   &state.TFState{},
	}
}

func (l *Loader) LoadFile(filename string) error {
	list, err := load(filename, l.Logger)
	if err != nil {
		return err
	}

	l.ListMap[filename] = list
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
		list, err := load(file, l.Logger)
		if err != nil {
			return err
		}
		filename := strings.Replace(file, modulePath, "", 1)
		fileKey := source + filename
		l.ListMap[fileKey] = list
	}

	return nil
}

func (l *Loader) LoadAllFile(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return err
	}
	filePattern := dir + "/*.tf"
	files, err := filepath.Glob(filePattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := l.LoadFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Loader) DumpFiles() map[string]*ast.ObjectList {
	return l.ListMap
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

func load(filename string, l *logger.Logger) (*ast.ObjectList, error) {
	l.Info(fmt.Sprintf("load `%s`", filename))
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("ERROR: Cannot open file %s", filename)
	}
	root, err := parser.Parse(b)
	if err != nil {
		l.Error(err)
		return nil, fmt.Errorf("ERROR: Parse error occurred by %s", filename)
	}

	list, _ := root.Node.(*ast.ObjectList)
	return list, nil
}
