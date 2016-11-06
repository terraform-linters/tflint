package loader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/logger"
)

type Loader struct {
	Config  *config.Config
	Logger  *logger.Logger
	ListMap map[string]*ast.ObjectList
}

func NewLoader(c *config.Config) *Loader {
	return &Loader{
		Config:  c,
		Logger:  logger.Init(c.Debug),
		ListMap: make(map[string]*ast.ObjectList),
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
		return errors.New(fmt.Sprintf("ERROR: module `%s` not found. Did you run `terraform get`?", source))
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

func load(filename string, l *logger.Logger) (*ast.ObjectList, error) {
	l.Info(fmt.Sprintf("load `%s`", filename))
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error(err)
		return nil, errors.New(fmt.Sprintf("ERROR: Cannot open file %s", filename))
	}
	root, err := parser.Parse(b)
	if err != nil {
		l.Error(err)
		return nil, errors.New(fmt.Sprintf("ERROR: Parse error occurred by %s", filename))
	}

	list, _ := root.Node.(*ast.ObjectList)
	return list, nil
}
