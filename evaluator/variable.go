package evaluator

import (
	"reflect"

	"github.com/hashicorp/hcl"
	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	hil_ast "github.com/hashicorp/hil/ast"
)

type hclVariable struct {
	Name         string `hcl:",key"`
	Default      interface{}
	Description  string
	DeclaredType string   `hcl:"type"`
	Fields       []string `hcl:",decodedFields"`
}

func detectVariables(listMap map[string]*hcl_ast.ObjectList) (map[string]hil_ast.Variable, error) {
	varMap := make(map[string]hil_ast.Variable)

	for _, list := range listMap {
		var variables []*hclVariable
		if err := hcl.DecodeObject(&variables, list.Filter("variable")); err != nil {
			return nil, err
		}

		for _, v := range variables {
			if v.Default == nil {
				continue
			}
			varName := "var." + v.Name
			varMap[varName] = parseVariable(v.Default, v.DeclaredType)
		}
	}

	return varMap, nil
}

func parseVariable(val interface{}, varType string) hil_ast.Variable {
	if varType == "" {
		switch reflect.TypeOf(val).Kind() {
		case reflect.String:
			varType = "string"
		case reflect.Slice:
			varType = "list"
		case reflect.Map:
			varType = "map"
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
	case "map":
		fallthrough
	case "list":
		s := reflect.ValueOf(val)

		switch reflect.TypeOf(s.Index(0).Interface()).Kind() {
		case reflect.Map:
			var variables map[string]hil_ast.Variable
			variables = map[string]hil_ast.Variable{}
			for i := 0; i < s.Len(); i++ {
				ms := reflect.ValueOf(s.Index(i).Interface())
				for _, k := range ms.MapKeys() {
					key := k.Interface().(string)
					value := ms.MapIndex(reflect.ValueOf(key)).Interface().(string)
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
