package evaluator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/schema"
)

func (e *Evaluator) initModule(module *schema.Module, c *config.Config) error {
	if err := module.Load(); err != nil {
		return err
	}

	varMap := make(map[string]ast.Variable)
	for k := range module.Attrs {
		if k != "source" {
			if varToken, ok := module.GetToken(k); ok {
				varName := "var." + k
				ev, err := e.evalModuleAttr(k, strings.Replace(varToken.Text, "\"", "", -1))
				if err != nil {
					return errors.New(fmt.Sprintf("Evaluation error: %s in %s:%d", err, varToken.Pos.Filename, varToken.Pos.Line))
				}
				varMap[varName] = parseVariable(ev, "")
			}
		}
	}

	module.EvalConfig = hil.EvalConfig{
		GlobalScope: &ast.BasicScope{
			VarMap: varMap,
		},
	}

	return nil
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
