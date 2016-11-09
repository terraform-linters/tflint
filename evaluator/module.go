package evaluator

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl"
	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hil_ast "github.com/hashicorp/hil/ast"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/loader"
)

type hclModule struct {
	Name    string
	Source  string
	Config  hil.EvalConfig
	ListMap map[string]*hcl_ast.ObjectList
}

func detectModules(listMap map[string]*hcl_ast.ObjectList, c *config.Config) (map[string]*hclModule, error) {
	moduleMap := make(map[string]*hclModule)

	for file, list := range listMap {
		for _, item := range list.Filter("module").Items {
			name, ok := item.Keys[0].Token.Value().(string)
			if !ok {
				return nil, errors.New(fmt.Sprintf("ERROR: Invalid module syntax in %s", file))
			}
			var module map[string]interface{}
			if err := hcl.DecodeObject(&module, item.Val); err != nil {
				return nil, err
			}

			moduleSource, ok := module["source"].(string)
			if !ok {
				return nil, errors.New(fmt.Sprintf("ERROR: Invalid module source in %s", name))
			}
			moduleKey := moduleKey(name, moduleSource)
			load := loader.NewLoader(c)
			err := load.LoadModuleFile(moduleKey, moduleSource)
			if err != nil {
				return nil, err
			}
			delete(module, "source")

			varMap := make(map[string]hil_ast.Variable)
			for k, v := range module {
				varName := "var." + k
				varMap[varName] = parseVariable(v, "")
			}

			moduleMap[moduleKey] = &hclModule{
				Name:   name,
				Source: moduleSource,
				Config: hil.EvalConfig{
					GlobalScope: &hil_ast.BasicScope{
						VarMap: varMap,
					},
				},
				ListMap: load.ListMap,
			}
		}
	}

	return moduleMap, nil
}

func moduleKey(name string, source string) string {
	base := "root." + name + "-" + source
	sum := md5.Sum([]byte(base))
	return hex.EncodeToString(sum[:])
}
