package evaluator

import (
	"reflect"
	"testing"

	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/k0kubun/pp"
)

func TestDetectVariables(t *testing.T) {
	cases := []struct {
		Name   string
		Input  map[string]string
		Result map[string]hilast.Variable
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
			Result: map[string]hilast.Variable{
				"var.type": hilast.Variable{
					Type:  hilast.TypeString,
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
			Result: map[string]hilast.Variable{},
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
			Result: map[string]hilast.Variable{},
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
			Result: map[string]hilast.Variable{
				"var.type": hilast.Variable{
					Type:  hilast.TypeString,
					Value: "name",
				},
				"var.stat": hilast.Variable{
					Type:  hilast.TypeString,
					Value: "usage",
				},
			},
			Error: false,
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*hclast.ObjectList)
		for k, v := range tc.Input {
			root, _ := parser.Parse([]byte(v))
			list, _ := root.Node.(*hclast.ObjectList)
			listMap[k] = list
		}

		result, err := detectVariables(listMap, map[string]*hclast.File{})
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestParseVariable(t *testing.T) {
	type Input struct {
		Val  interface{}
		Type string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result hilast.Variable
	}{
		{
			Name: "parse string with correct type",
			Input: Input{
				Val:  "test",
				Type: "string",
			},
			Result: hilast.Variable{
				Type:  hilast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with correct type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "list",
			},
			Result: hilast.Variable{
				Type: hilast.TypeList,
				Value: []hilast.Variable{
					hilast.Variable{
						Type:  hilast.TypeString,
						Value: "test1",
					},
					hilast.Variable{
						Type:  hilast.TypeString,
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
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"test1": hilast.Variable{
						Type:  hilast.TypeString,
						Value: "test2",
					},
					"test3": hilast.Variable{
						Type:  hilast.TypeString,
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
			Result: hilast.Variable{
				Type:  hilast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with unknown type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "",
			},
			Result: hilast.Variable{
				Type: hilast.TypeList,
				Value: []hilast.Variable{
					hilast.Variable{
						Type:  hilast.TypeString,
						Value: "test1",
					},
					hilast.Variable{
						Type:  hilast.TypeString,
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
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"test1": hilast.Variable{
						Type:  hilast.TypeString,
						Value: "test2",
					},
					"test3": hilast.Variable{
						Type:  hilast.TypeString,
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
			Result: hilast.Variable{
				Type:  hilast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with incorrect type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "string",
			},
			Result: hilast.Variable{
				Type: hilast.TypeList,
				Value: []hilast.Variable{
					hilast.Variable{
						Type:  hilast.TypeString,
						Value: "test1",
					},
					hilast.Variable{
						Type:  hilast.TypeString,
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
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"test1": hilast.Variable{
						Type:  hilast.TypeString,
						Value: "test2",
					},
					"test3": hilast.Variable{
						Type:  hilast.TypeString,
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
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"a": hilast.Variable{
						Type: hilast.TypeList,
						Value: []hilast.Variable{
							hilast.Variable{
								Type:  hilast.TypeString,
								Value: "test1",
							},
							hilast.Variable{
								Type:  hilast.TypeString,
								Value: "test2",
							},
						},
					},
					"b": hilast.Variable{
						Type: hilast.TypeMap,
						Value: map[string]hilast.Variable{
							"test3": hilast.Variable{
								Type:  hilast.TypeString,
								Value: 1,
							},
							"test4": hilast.Variable{
								Type:  hilast.TypeString,
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
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}
