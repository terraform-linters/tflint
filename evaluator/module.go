package evaluator

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"
	"strings"

	"reflect"

	"github.com/hashicorp/hcl"
	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/loader"
	"github.com/wata727/tflint/schema"
)

type hclModule struct {
	Name       string
	Source     string
	ObjectItem *hclast.ObjectItem
	File       string
	Config     hil.EvalConfig
	Templates  map[string]*hclast.File
	Schema     []*schema.Template
}

func (e *Evaluator) initModule(module *schema.Module, c *config.Config) error {
	if err := module.Load(); err != nil {
		return err
	}

	varMap := make(map[string]hilast.Variable)
	for k := range module.Attrs {
		if varToken, ok := module.GetToken(k); ok {
			varName := "var." + k
			ev, err := e.evalModuleAttr(k, strings.Replace(varToken.Text, "\"", "", -1))
			if err != nil {
				return err
			}
			varMap[varName] = parseVariable(ev, "")
		}
	}

	module.EvalConfig = hil.EvalConfig{
		GlobalScope: &hilast.BasicScope{
			VarMap: varMap,
		},
	}

	return nil
}

func (e *Evaluator) detectModules(templates map[string]*hclast.File, c *config.Config) (map[string]*hclModule, error) {
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
			if err := load.LoadModuleFile(moduleKey, moduleSource); err != nil {
				return nil, err
			}
			schema, err := schema.Make(load.Files)
			if err != nil {
				return nil, err
			}
			delete(module, "source")

			varMap := make(map[string]hilast.Variable)
			for k, v := range module {
				varName := "var." + k
				ev, err := e.evalModuleAttr(k, v)
				if err != nil {
					return nil, err
				}
				varMap[varName] = parseVariable(ev, "")
			}

			moduleMap[moduleKey] = &hclModule{
				Name:       name,
				Source:     moduleSource,
				ObjectItem: item,
				File:       file,
				Config: hil.EvalConfig{
					GlobalScope: &hilast.BasicScope{
						VarMap: varMap,
					},
				},
				Templates: load.Templates,
				Schema:    schema,
			}
		}
	}

	return moduleMap, nil
}

func (e *Evaluator) evalModuleAttr(key string, val interface{}) (interface{}, error) {
	if v, ok := val.(string); ok {
		ev, err := e.Eval(v)
		if err != nil {
			return nil, err
		}
		if estr, ok := ev.(string); ok && estr == "[NOT EVALUABLE]" {
			ev = v
		}

		// In parseVariable function, map is expected to be in slice.
		switch reflect.ValueOf(ev).Kind() {
		case reflect.Map:
			return []interface{}{ev}, nil
		default:
			return ev, nil
		}
	}
	return val, nil
}

func moduleKey(name string, source string) string {
	base := "root." + name + "-" + source
	sum := md5.Sum([]byte(base)) // #nosec
	return hex.EncodeToString(sum[:])
}
