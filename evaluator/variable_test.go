package evaluator

import (
	"reflect"
	"testing"

	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	hil_ast "github.com/hashicorp/hil/ast"
)

func TestDetectVariable(t *testing.T) {
	cases := []struct {
		Name   string
		Input  map[string]string
		Result map[string]hil_ast.Variable
		Error  bool
	}{
		{
			Name: "return hil variable from hcl object list",
			Input: map[string]string{
				"variable.tf": `
variable "type" {
	default = "name"
}`,
			},
			Result: map[string]hil_ast.Variable{
				"var.type": hil_ast.Variable{
					Type:  hil_ast.TypeString,
					Value: "name",
				},
			},
			Error: false,
		},
		{
			Name: "return empty from hcl object list when default is not found",
			Input: map[string]string{
				"variable.tf": `variable "type" {}`,
			},
			Result: map[string]hil_ast.Variable{},
			Error:  false,
		},
		{
			Name: "return empty from hcl object list when variable not found",
			Input: map[string]string{
				"variable.tf": `
provider "aws" {
	region = "us-east-1"
}`,
			},
			Result: map[string]hil_ast.Variable{},
			Error:  false,
		},
		{
			Name: "return hil variables from multi hcl object list",
			Input: map[string]string{
				"variable1.tf": `
variable "type" {
	default = "name"
}`,
				"variable2.tf": `
variable "stat" {
	default = "usage"
}`,
			},
			Result: map[string]hil_ast.Variable{
				"var.type": hil_ast.Variable{
					Type:  hil_ast.TypeString,
					Value: "name",
				},
				"var.stat": hil_ast.Variable{
					Type:  hil_ast.TypeString,
					Value: "usage",
				},
			},
			Error: false,
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*hcl_ast.ObjectList)
		for k, v := range tc.Input {
			root, _ := parser.Parse([]byte(v))
			list, _ := root.Node.(*hcl_ast.ObjectList)
			listMap[k] = list
		}

		result, err := detectVariables(listMap)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

type Input struct {
	Val  interface{}
	Type string
}

func TestParseVariable(t *testing.T) {
	cases := []struct {
		Name   string
		Input  Input
		Result hil_ast.Variable
	}{
		{
			Name: "parse string with correct type",
			Input: Input{
				Val:  "test",
				Type: "string",
			},
			Result: hil_ast.Variable{
				Type:  hil_ast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with correct type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "list",
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeList,
				Value: []hil_ast.Variable{
					hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test1",
					},
					hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test2",
					},
				},
			},
		},
		{
			Name: "parse map with correct type",
			Input: Input{
				// HCL map variable is map in slice in Go.
				Val:  []map[string]string{map[string]string{"test1": "test2", "test3": "test4"}},
				Type: "map",
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeMap,
				Value: map[string]hil_ast.Variable{
					"test1": hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test2",
					},
					"test3": hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test4",
					},
				},
			},
		},
		{
			Name: "parse string with unknown type",
			Input: Input{
				Val:  "test",
				Type: "",
			},
			Result: hil_ast.Variable{
				Type:  hil_ast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with unknown type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "",
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeList,
				Value: []hil_ast.Variable{
					hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test1",
					},
					hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test2",
					},
				},
			},
		},
		{
			Name: "parse map with unknown type",
			Input: Input{
				Val:  []map[string]string{map[string]string{"test1": "test2", "test3": "test4"}},
				Type: "",
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeMap,
				Value: map[string]hil_ast.Variable{
					"test1": hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test2",
					},
					"test3": hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test4",
					},
				},
			},
		},
		{
			Name: "parse string with incorrect type",
			Input: Input{
				Val:  "test",
				Type: "map",
			},
			Result: hil_ast.Variable{
				Type:  hil_ast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with incorrect type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "string",
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeList,
				Value: []hil_ast.Variable{
					hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test1",
					},
					hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test2",
					},
				},
			},
		},
		{
			Name: "parse map with incorrect type",
			Input: Input{
				Val:  []map[string]string{map[string]string{"test1": "test2", "test3": "test4"}},
				Type: "slice",
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeMap,
				Value: map[string]hil_ast.Variable{
					"test1": hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test2",
					},
					"test3": hil_ast.Variable{
						Type:  hil_ast.TypeString,
						Value: "test4",
					},
				},
			},
		},
		{
			Name: "parse complex struct var",
			// ```HCL
			// value = {
			//     a = ["test1", "test2"]
			//     b = {
			//         test3 = 1
			//         test4 = 10
			//     }
			// }
			// ```
			Input: Input{
				Val: []map[string]interface{}{
					map[string]interface{}{
						"a": []string{"test1", "test2"},
						"b": []map[string]int{
							map[string]int{
								"test3": 1,
								"test4": 10,
							},
						},
					},
				},
			},
			Result: hil_ast.Variable{
				Type: hil_ast.TypeMap,
				Value: map[string]hil_ast.Variable{
					"a": hil_ast.Variable{
						Type: hil_ast.TypeList,
						Value: []hil_ast.Variable{
							hil_ast.Variable{
								Type:  hil_ast.TypeString,
								Value: "test1",
							},
							hil_ast.Variable{
								Type:  hil_ast.TypeString,
								Value: "test2",
							},
						},
					},
					"b": hil_ast.Variable{
						Type: hil_ast.TypeMap,
						Value: map[string]hil_ast.Variable{
							"test3": hil_ast.Variable{
								Type:  hil_ast.TypeString,
								Value: 1,
							},
							"test4": hil_ast.Variable{
								Type:  hil_ast.TypeString,
								Value: 10,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		result := parseVariable(tc.Input.Val, tc.Input.Type)
		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}
