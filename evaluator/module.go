package evaluator

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/hcl"
	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/loader"
)

type hclModule struct {
	Name    string
	Source  string
	Config  hil.EvalConfig
	ListMap map[string]*hclast.ObjectList
}

func detectModules(listMap map[string]*hclast.ObjectList, c *config.Config) (map[string]*hclModule, error) {
	moduleMap := make(map[string]*hclModule)

	for file, list := range listMap {
		for _, item := range list.Filter("module").Items {
			name, ok := item.Keys[0].Token.Value().(string)
			if !ok {
				return nil, fmt.Errorf("ERROR: Invalid module syntax in %s", file)
			}
			var module map[string]interface{}
			if err := hcl.DecodeObject(&module, item.Val); err != nil {
				return nil, err
			}

			moduleSource, ok := module["source"].(string)
			if !ok {
				return nil, fmt.Errorf("ERROR: Invalid module source in %s", name)
			}
			moduleKey := moduleKey(name, moduleSource)
			load := loader.NewLoader(c.Debug)
			err := load.LoadModuleFile(moduleKey, moduleSource)
			if err != nil {
				return nil, err
			}
			delete(module, "source")

			varMap := make(map[string]hilast.Variable)
			for k, v := range module {
				varName := "var." + k
				varMap[varName] = parseVariable(v, "")
			}

			moduleMap[moduleKey] = &hclModule{
				Name:   name,
				Source: moduleSource,
				Config: hil.EvalConfig{
					GlobalScope: &hilast.BasicScope{
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
	sum := md5.Sum([]byte(base)) // #nosec
	return hex.EncodeToString(sum[:])
}
