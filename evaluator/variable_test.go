package evaluator

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hil/ast"
)

type Input struct {
	Val  interface{}
	Type string
}

func TestParseVariable(t *testing.T) {
	cases := []struct {
		Name   string
		Input  Input
		Result ast.Variable
	}{
		{
			Name: "parse string with correct type",
			Input: Input{
				Val:  "test",
				Type: "string",
			},
			Result: ast.Variable{
				Type:  ast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with correct type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "list",
			},
			Result: ast.Variable{
				Type: ast.TypeList,
				Value: []ast.Variable{
					ast.Variable{
						Type:  ast.TypeString,
						Value: "test1",
					},
					ast.Variable{
						Type:  ast.TypeString,
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
			Result: ast.Variable{
				Type: ast.TypeMap,
				Value: map[string]ast.Variable{
					"test1": ast.Variable{
						Type:  ast.TypeString,
						Value: "test2",
					},
					"test3": ast.Variable{
						Type:  ast.TypeString,
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
			Result: ast.Variable{
				Type:  ast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with unknown type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "",
			},
			Result: ast.Variable{
				Type: ast.TypeList,
				Value: []ast.Variable{
					ast.Variable{
						Type:  ast.TypeString,
						Value: "test1",
					},
					ast.Variable{
						Type:  ast.TypeString,
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
			Result: ast.Variable{
				Type: ast.TypeMap,
				Value: map[string]ast.Variable{
					"test1": ast.Variable{
						Type:  ast.TypeString,
						Value: "test2",
					},
					"test3": ast.Variable{
						Type:  ast.TypeString,
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
			Result: ast.Variable{
				Type:  ast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with incorrect type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "string",
			},
			Result: ast.Variable{
				Type: ast.TypeList,
				Value: []ast.Variable{
					ast.Variable{
						Type:  ast.TypeString,
						Value: "test1",
					},
					ast.Variable{
						Type:  ast.TypeString,
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
			Result: ast.Variable{
				Type: ast.TypeMap,
				Value: map[string]ast.Variable{
					"test1": ast.Variable{
						Type:  ast.TypeString,
						Value: "test2",
					},
					"test3": ast.Variable{
						Type:  ast.TypeString,
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
			Result: ast.Variable{
				Type: ast.TypeMap,
				Value: map[string]ast.Variable{
					"a": ast.Variable{
						Type: ast.TypeList,
						Value: []ast.Variable{
							ast.Variable{
								Type:  ast.TypeString,
								Value: "test1",
							},
							ast.Variable{
								Type:  ast.TypeString,
								Value: "test2",
							},
						},
					},
					"b": ast.Variable{
						Type: ast.TypeMap,
						Value: map[string]ast.Variable{
							"test3": ast.Variable{
								Type:  ast.TypeString,
								Value: 1,
							},
							"test4": ast.Variable{
								Type:  ast.TypeString,
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
