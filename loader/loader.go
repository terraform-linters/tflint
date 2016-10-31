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
)

func LoadFile(listmap map[string]*ast.ObjectList, filename string) (map[string]*ast.ObjectList, error) {
	if listmap == nil {
		listmap = make(map[string]*ast.ObjectList)
	}

	list, err := load(filename)
	if err != nil {
		return nil, err
	}

	listmap[filename] = list
	return listmap, nil
}

func LoadModuleFile(moduleKey string, source string) (map[string]*ast.ObjectList, error) {
	var listmap = make(map[string]*ast.ObjectList)

	modulePath := ".terraform/modules/" + moduleKey
	if _, err := os.Stat(modulePath); err != nil {
		return nil, err
	}
	filePattern := modulePath + "/*.tf"
	files, err := filepath.Glob(filePattern)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		list, err := load(file)
		if err != nil {
			return nil, err
		}
		filename := strings.Replace(file, modulePath, "", 1)
		fileKey := source + filename
		listmap[fileKey] = list
	}

	return listmap, nil
}

func LoadAllFile(dir string) (map[string]*ast.ObjectList, error) {
	var listmap = make(map[string]*ast.ObjectList)

	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}
	filePattern := dir + "/*.tf"
	files, err := filepath.Glob(filePattern)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		listmap, err = LoadFile(listmap, file)
		if err != nil {
			return nil, err
		}
	}

	return listmap, nil
}

func load(filename string) (*ast.ObjectList, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ERROR: Cannot open file %s\n", filename))
	}
	root, err := parser.Parse(b)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ERROR: Parse error occurred by %s\n", filename))
	}

	list, _ := root.Node.(*ast.ObjectList)
	return list, nil
}
