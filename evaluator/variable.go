package evaluator

import (
	"os"
	"reflect"
	"strings"

	"fmt"

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
		envVars, err := decodeEnvVars(variables)
		if err != nil {
			return nil, err
		}
		tfvars, err := decodeTFVars(varfile)
		if err != nil {
			return nil, err
		}

		for _, v := range variables {
			overriddenVariable(v, envVars)
			overriddenVariable(v, tfvars)
			if v.Default == nil {
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

func decodeEnvVars(variables []*hclVariable) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, e := range os.Environ() {
		idx := strings.Index(e, "=")
		envKey := e[:idx]
		envVal := e[idx+1:]

		if strings.HasPrefix(envKey, "TF_VAR_") {
			varName := strings.Replace(envKey, "TF_VAR_", "", 1)
			for _, v := range variables {
				if v.Name != varName {
					continue
				}

				var varType string
				if v.DeclaredType == "" {
					varType = deduceType(v.Default)
				} else {
					varType = v.DeclaredType
				}

				if varType == HCL_STRING_VARTYPE {
					envVal = fmt.Sprintf("\"%s\"", envVal)
				}

				var r map[string]interface{}
				if err := hcl.Decode(&r, fmt.Sprintf("%s = %s", varName, envVal)); err != nil {
					return nil, err
				}
				result = append(result, r)
			}
		}
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

func deduceType(val interface{}) string {
	if val == nil {
		return HCL_STRING_VARTYPE
	}

	var varType string

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

	return varType
}

func parseVariable(val interface{}, varType string) hilast.Variable {
	// varType is overwrite invariably. Because, happen panic when used in incorrect type
	varType = deduceType(val)

	var hilVar hilast.Variable
	switch varType {
	case HCL_STRING_VARTYPE:
		hilVar = hilast.Variable{
			Type:  hilast.TypeString,
			Value: fmt.Sprint(val),
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
		//     {
		//         "key": []map[string]string{
		//             {
		//                 "name":  "test",
		//                 "value": "hcl",
		//             },
		//         },
		//     },
		// }
		//
		fallthrough
	case HCL_LIST_VARTYPE:
		s := reflect.ValueOf(val)

		if s.Len() == 0 {
			hilVar = hilast.Variable{
				Type:  hilast.TypeList,
				Value: []hilast.Variable{},
			}
		} else {
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
	}

	return hilVar
}
