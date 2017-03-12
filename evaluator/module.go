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
	Name      string
	Source    string
	Config    hil.EvalConfig
	Templates map[string]*hclast.File
}

func detectModules(templates map[string]*hclast.File, c *config.Config) (map[string]*hclModule, error) {
	moduleMap := make(map[string]*hclModule)

	for file, template := range templates {
		for _, item := range template.Node.(*hclast.ObjectList).Filter("module").Items {
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
				Templates: load.Templates,
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
