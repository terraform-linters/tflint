package evaluator

import (
	"crypto/md5"
	"encoding/hex"
	"errors"

	"github.com/hashicorp/hcl"
	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hil_ast "github.com/hashicorp/hil/ast"
	"github.com/wata727/tflint/loader"
)

type hclModule struct {
	Config  hil.EvalConfig
	ListMap map[string]*hcl_ast.ObjectList
}

func detectModules(listMap map[string]*hcl_ast.ObjectList) (map[string]*hclModule, error) {
	moduleMap := make(map[string]*hclModule)

	for _, list := range listMap {
		for _, item := range list.Filter("module").Items {
			name, ok := item.Keys[0].Token.Value().(string)
			if !ok {
				return nil, errors.New("invalid module name")
			}
			var module map[string]interface{}
			if err := hcl.DecodeObject(&module, item.Val); err != nil {
				return nil, err
			}

			moduleSource, ok := module["source"].(string)
			if !ok {
				return nil, errors.New("invalid module source")
			}
			moduleKey := moduleKey(name, moduleSource)
			moduleListMap, err := loader.LoadModuleFile(moduleKey, moduleSource)
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
				Config: hil.EvalConfig{
					GlobalScope: &hil_ast.BasicScope{
						VarMap: varMap,
					},
				},
				ListMap: moduleListMap,
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
