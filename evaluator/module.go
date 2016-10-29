package evaluator

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

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

func detectModules(listmap map[string]*hcl_ast.ObjectList) (map[string]*hclModule, error) {
	modulemap := make(map[string]*hclModule)

	for _, list := range listmap {
		for _, item := range list.Filter("module").Items {
			name := item.Keys[0].Token.Value().(string)
			var module map[string]interface{}
			if err := hcl.DecodeObject(&module, item.Val); err != nil {
				return nil, err
			}

			moduleSource := fmt.Sprint(module["source"])
			moduleKey := moduleKey(name, moduleSource)
			moduleListmap, err := loader.LoadModuleFile(moduleKey, moduleSource)
			if err != nil {
				return nil, err
			}
			delete(module, "source")

			varmap := make(map[string]hil_ast.Variable)
			for k, v := range module {
				varName := "var." + k
				varmap[varName] = parseVariable(v, "")
			}

			modulemap[moduleKey] = &hclModule{
				Config: hil.EvalConfig{
					GlobalScope: &hil_ast.BasicScope{
						VarMap: varmap,
					},
				},
				ListMap: moduleListmap,
			}
		}
	}

	return modulemap, nil
}

func moduleKey(name string, source string) string {
	base := "root." + name + "-" + source
	sum := md5.Sum([]byte(base))
	return hex.EncodeToString(sum[:])
}
