package evaluator

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl"
	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hil_ast "github.com/hashicorp/hil/ast"
)

type Evaluator struct {
	Config hil.EvalConfig
}

type hclVariable struct {
	Name         string `hcl:",key"`
	Default      interface{}
	Description  string
	DeclaredType string   `hcl:"type"`
	Fields       []string `hcl:",decodedFields"`
}

func NewEvaluator(listmap map[string]*hcl_ast.ObjectList) *Evaluator {
	varmap := detectVariables(listmap)

	evaluator := &Evaluator{
		Config: hil.EvalConfig{
			GlobalScope: &hil_ast.BasicScope{
				VarMap: varmap,
			},
		},
	}

	return evaluator
}

func detectVariables(listmap map[string]*hcl_ast.ObjectList) map[string]hil_ast.Variable {
	varmap := make(map[string]hil_ast.Variable)

	for _, list := range listmap {
		var variables []*hclVariable
		if err := hcl.DecodeObject(&variables, list.Filter("variable")); err != nil {
			return varmap
		}

		for _, v := range variables {
			varName := "var." + v.Name
			varmap[varName] = parseVariable(v.Default, v.DeclaredType)
		}
	}

	return varmap
}

func parseVariable(val interface{}, varType string) hil_ast.Variable {
	if varType == "" {
		switch reflect.TypeOf(val).Kind() {
		case reflect.String:
			varType = "string"
		case reflect.Slice:
			varType = "list"
		case reflect.Map:
			varType = "list"
		default:
			varType = "string"
		}
	}

	var hilVar hil_ast.Variable
	switch varType {
	case "string":
		hilVar = hil_ast.Variable{
			Type:  hil_ast.TypeString,
			Value: val,
		}
	case "list":
		s := reflect.ValueOf(val)

		switch reflect.TypeOf(s.Index(0).Interface()).Kind() {
		case reflect.Map:
			var variables map[string]hil_ast.Variable
			variables = map[string]hil_ast.Variable{}
			for i := 0; i < s.Len(); i++ {
				ms := reflect.ValueOf(s.Index(i).Interface())
				for _, k := range ms.MapKeys() {
					key := fmt.Sprint(k.Interface())
					value := fmt.Sprint(ms.MapIndex(reflect.ValueOf(key)).Interface())
					variables[key] = parseVariable(value, "")
				}
			}
			hilVar = hil_ast.Variable{
				Type:  hil_ast.TypeMap,
				Value: variables,
			}
		default:
			var variables []hil_ast.Variable
			for i := 0; i < s.Len(); i++ {
				variables = append(variables, parseVariable(s.Index(i).Interface(), ""))
			}
			hilVar = hil_ast.Variable{
				Type:  hil_ast.TypeList,
				Value: variables,
			}
		}
	}

	return hilVar
}

func (e *Evaluator) Eval(src string) string {
	root, _ := hil.Parse(src)
	result, _ := hil.Eval(root, &e.Config)
	return result.Value.(string)
}
