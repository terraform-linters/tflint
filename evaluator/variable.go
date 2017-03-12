package evaluator

import (
	"reflect"

	"github.com/hashicorp/hcl"
	hclast "github.com/hashicorp/hcl/hcl/ast"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/hashicorp/terraform/helper/variables"
)

type hclVariable struct {
	Name         string `hcl:",key"`
	Default      interface{}
	Description  string
	DeclaredType string   `hcl:"type"`
	Fields       []string `hcl:",decodedFields"`
}

const HCL_STRING_VARTYPE = "string"
const HCL_LIST_VARTYPE = "list"
const HCL_MAP_VARTYPE = "map"

func detectVariables(templates map[string]*hclast.File, varfile []*hclast.File) (map[string]hilast.Variable, error) {
	varMap := make(map[string]hilast.Variable)

	for _, template := range templates {
		var variables []*hclVariable
		if err := hcl.DecodeObject(&variables, template.Node.(*hclast.ObjectList).Filter("variable")); err != nil {
			return nil, err
		}
		tfvars, err := decodeTFVars(varfile)
		if err != nil {
			return nil, err
		}

		for _, v := range variables {
			if overriddenVariable(v, tfvars); v.Default == nil {
				continue
			}
			varName := "var." + v.Name
			varMap[varName] = parseVariable(v.Default, v.DeclaredType)
		}
	}

	return varMap, nil
}

func decodeTFVars(varfile []*hclast.File) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, vars := range varfile {
		var r map[string]interface{}
		if err := hcl.DecodeObject(&r, vars.Node); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func overriddenVariable(v *hclVariable, tfvars []map[string]interface{}) {
	for _, vars := range tfvars {
		val := vars[v.Name]
		if val == nil {
			continue
		}

		switch reflect.TypeOf(val).Kind() {
		case reflect.String:
			fallthrough
		case reflect.Slice:
			v.Default = val
		case reflect.Map:
			if d, ok := v.Default.(map[string]interface{}); ok {
				v.Default = variables.Merge(d, val.(map[string]interface{}))
			} else {
				v.Default = val
			}
		}
	}
}

func parseVariable(val interface{}, varType string) hilast.Variable {
	// varType is overwrite invariably. Because, happen panic when used in incorrect type
	switch reflect.TypeOf(val).Kind() {
	case reflect.String:
		varType = HCL_STRING_VARTYPE
	case reflect.Slice:
		varType = HCL_LIST_VARTYPE
	case reflect.Map:
		varType = HCL_MAP_VARTYPE
	default:
		varType = HCL_STRING_VARTYPE
	}

	var hilVar hilast.Variable
	switch varType {
	case HCL_STRING_VARTYPE:
		hilVar = hilast.Variable{
			Type:  hilast.TypeString,
			Value: val,
		}
	case HCL_MAP_VARTYPE:
		// When HCL map var convert(parse) to Go var,
		// get map in slice. following example:
		//
		// ```HCL
		// key = {
		//     name = "test"
		//     value = "hcl"
		// }
		// ```
		//
		// Incorrect:
		//
		// map[string]string{
		//     "key": map[string][string]{
		//         "name":  "test",
		//         "value": "hcl",
		//     },
		// }
		//
		// Correct:
		//
		// []map[string]string{
		//     map[string]string{
		//         "name":  "test",
		//         "value": "hcl",
		//     },
		// }
		//
		fallthrough
	case HCL_LIST_VARTYPE:
		s := reflect.ValueOf(val)

		switch reflect.TypeOf(s.Index(0).Interface()).Kind() {
		case reflect.Map:
			var variables map[string]hilast.Variable
			variables = map[string]hilast.Variable{}
			for i := 0; i < s.Len(); i++ {
				ms := reflect.ValueOf(s.Index(i).Interface())
				for _, k := range ms.MapKeys() {
					key := k.Interface().(string)
					value := ms.MapIndex(reflect.ValueOf(key)).Interface()
					variables[key] = parseVariable(value, "")
				}
			}
			hilVar = hilast.Variable{
				Type:  hilast.TypeMap,
				Value: variables,
			}
		default:
			var variables []hilast.Variable
			for i := 0; i < s.Len(); i++ {
				variables = append(variables, parseVariable(s.Index(i).Interface(), ""))
			}
			hilVar = hilast.Variable{
				Type:  hilast.TypeList,
				Value: variables,
			}
		}
	}

	return hilVar
}
