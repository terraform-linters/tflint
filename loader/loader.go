package loader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
)

func LoadFile(listmap map[string]*ast.ObjectList, filename string) (map[string]*ast.ObjectList, error) {
	if listmap == nil {
		listmap = make(map[string]*ast.ObjectList)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ERROR: Cannot open file %s\n", filename))
	}
	root, err := parser.Parse(b)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ERROR: Parse error occurred by %s\n", filename))
	}

	list, _ := root.Node.(*ast.ObjectList)
	listmap[filename] = list
	return listmap, nil
}

func LoadAllFile() (map[string]*ast.ObjectList, error) {
	var listmap = make(map[string]*ast.ObjectList)

	files, err := filepath.Glob("./*.tf")
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
