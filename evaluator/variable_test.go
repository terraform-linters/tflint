package evaluator

import (
	"reflect"
	"testing"

	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	hil_ast "github.com/hashicorp/hil/ast"
)

func TestDetectVariable(t *testing.T) {
	cases := []struct {
		Name   string
		Input  map[string]*hcl_ast.ObjectList
		Result map[string]hil_ast.Variable
		Error  bool
	}{
		{
			Name: "return hil variable from hcl object list",
			Input: map[string]*hcl_ast.ObjectList{
				"variable.tf": &hcl_ast.ObjectList{
					Items: []*hcl_ast.ObjectItem{
						&hcl_ast.ObjectItem{
							Keys: []*hcl_ast.ObjectKey{
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 4,
										Pos: token.Pos{
											Filename: "",
											Offset:   0,
											Line:     1,
											Column:   1,
										},
										Text: "variable",
										JSON: false,
									},
								},
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 9,
										Pos: token.Pos{
											Filename: "",
											Offset:   9,
											Line:     1,
											Column:   10,
										},
										Text: "\"type\"",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{
								Filename: "",
								Offset:   0,
								Line:     0,
								Column:   0,
							},
							Val: &hcl_ast.ObjectType{
								Lbrace: token.Pos{
									Filename: "",
									Offset:   16,
									Line:     1,
									Column:   17,
								},
								Rbrace: token.Pos{
									Filename: "",
									Offset:   96,
									Line:     6,
									Column:   1,
								},
								List: &hcl_ast.ObjectList{
									Items: []*hcl_ast.ObjectItem{
										&hcl_ast.ObjectItem{
											Keys: []*hcl_ast.ObjectKey{
												&hcl_ast.ObjectKey{
													Token: token.Token{
														Type: 4,
														Pos: token.Pos{
															Filename: "",
															Offset:   22,
															Line:     2,
															Column:   5,
														},
														Text: "default",
														JSON: false,
													},
												},
											},
											Assign: token.Pos{
												Filename: "",
												Offset:   34,
												Line:     2,
												Column:   17,
											},
											Val: &hcl_ast.LiteralType{
												Token: token.Token{
													Type: 9,
													Pos: token.Pos{
														Filename: "",
														Offset:   36,
														Line:     2,
														Column:   19,
													},
													Text: "\"name\"",
													JSON: false,
												},
												LineComment: (*hcl_ast.CommentGroup)(nil),
											},
											LeadComment: (*hcl_ast.CommentGroup)(nil),
											LineComment: (*hcl_ast.CommentGroup)(nil),
										},
									},
								},
							},
						},
					},
				},
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
			Input: map[string]*hcl_ast.ObjectList{
				"variable.tf": &hcl_ast.ObjectList{
					Items: []*hcl_ast.ObjectItem{
						&hcl_ast.ObjectItem{
							Keys: []*hcl_ast.ObjectKey{
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 4,
										Pos: token.Pos{
											Filename: "",
											Offset:   0,
											Line:     1,
											Column:   1,
										},
										Text: "variable",
										JSON: false,
									},
								},
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 9,
										Pos: token.Pos{
											Filename: "",
											Offset:   9,
											Line:     1,
											Column:   10,
										},
										Text: "\"type\"",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{
								Filename: "",
								Offset:   0,
								Line:     0,
								Column:   0,
							},
							Val: &hcl_ast.ObjectType{
								Lbrace: token.Pos{
									Filename: "",
									Offset:   16,
									Line:     1,
									Column:   17,
								},
								Rbrace: token.Pos{
									Filename: "",
									Offset:   96,
									Line:     6,
									Column:   1,
								},
								List: &hcl_ast.ObjectList{
									Items: []*hcl_ast.ObjectItem{},
								},
							},
						},
					},
				},
			},
			Result: map[string]hil_ast.Variable{},
			Error:  false,
		},
		{
			Name: "return empty from hcl object list when variable not found",
			Input: map[string]*hcl_ast.ObjectList{
				"template.tf": &hcl_ast.ObjectList{
					Items: []*hcl_ast.ObjectItem{
						&hcl_ast.ObjectItem{
							Keys: []*hcl_ast.ObjectKey{
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 4,
										Pos: token.Pos{
											Filename: "",
											Offset:   0,
											Line:     1,
											Column:   1,
										},
										Text: "provider",
										JSON: false,
									},
								},
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 9,
										Pos: token.Pos{
											Filename: "",
											Offset:   9,
											Line:     1,
											Column:   10,
										},
										Text: "\"aws\"",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{
								Filename: "",
								Offset:   0,
								Line:     0,
								Column:   0,
							},
							Val: &hcl_ast.ObjectType{
								Lbrace: token.Pos{
									Filename: "",
									Offset:   15,
									Line:     1,
									Column:   16,
								},
								Rbrace: token.Pos{
									Filename: "",
									Offset:   40,
									Line:     3,
									Column:   1,
								},
								List: &hcl_ast.ObjectList{
									Items: []*hcl_ast.ObjectItem{
										&hcl_ast.ObjectItem{
											Keys: []*hcl_ast.ObjectKey{
												&hcl_ast.ObjectKey{
													Token: token.Token{
														Type: 4,
														Pos: token.Pos{
															Filename: "",
															Offset:   19,
															Line:     2,
															Column:   3,
														},
														Text: "region",
														JSON: false,
													},
												},
											},
											Assign: token.Pos{
												Filename: "",
												Offset:   26,
												Line:     2,
												Column:   10,
											},
											Val: &hcl_ast.LiteralType{
												Token: token.Token{
													Type: 9,
													Pos: token.Pos{
														Filename: "",
														Offset:   28,
														Line:     2,
														Column:   12,
													},
													Text: "\"us-east-1\"",
													JSON: false,
												},
												LineComment: (*hcl_ast.CommentGroup)(nil),
											},
											LeadComment: (*hcl_ast.CommentGroup)(nil),
											LineComment: (*hcl_ast.CommentGroup)(nil),
										},
									},
								},
							},
						},
					},
				},
			},
			Result: map[string]hil_ast.Variable{},
			Error:  false,
		},
		{
			Name: "return hil variables from multi hcl object list",
			Input: map[string]*hcl_ast.ObjectList{
				"variable1.tf": &hcl_ast.ObjectList{
					Items: []*hcl_ast.ObjectItem{
						&hcl_ast.ObjectItem{
							Keys: []*hcl_ast.ObjectKey{
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 4,
										Pos: token.Pos{
											Filename: "",
											Offset:   0,
											Line:     1,
											Column:   1,
										},
										Text: "variable",
										JSON: false,
									},
								},
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 9,
										Pos: token.Pos{
											Filename: "",
											Offset:   9,
											Line:     1,
											Column:   10,
										},
										Text: "\"type\"",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{
								Filename: "",
								Offset:   0,
								Line:     0,
								Column:   0,
							},
							Val: &hcl_ast.ObjectType{
								Lbrace: token.Pos{
									Filename: "",
									Offset:   16,
									Line:     1,
									Column:   17,
								},
								Rbrace: token.Pos{
									Filename: "",
									Offset:   96,
									Line:     6,
									Column:   1,
								},
								List: &hcl_ast.ObjectList{
									Items: []*hcl_ast.ObjectItem{
										&hcl_ast.ObjectItem{
											Keys: []*hcl_ast.ObjectKey{
												&hcl_ast.ObjectKey{
													Token: token.Token{
														Type: 4,
														Pos: token.Pos{
															Filename: "",
															Offset:   22,
															Line:     2,
															Column:   5,
														},
														Text: "default",
														JSON: false,
													},
												},
											},
											Assign: token.Pos{
												Filename: "",
												Offset:   34,
												Line:     2,
												Column:   17,
											},
											Val: &hcl_ast.LiteralType{
												Token: token.Token{
													Type: 9,
													Pos: token.Pos{
														Filename: "",
														Offset:   36,
														Line:     2,
														Column:   19,
													},
													Text: "\"name\"",
													JSON: false,
												},
												LineComment: (*hcl_ast.CommentGroup)(nil),
											},
											LeadComment: (*hcl_ast.CommentGroup)(nil),
											LineComment: (*hcl_ast.CommentGroup)(nil),
										},
									},
								},
							},
						},
					},
				},
				"variable2.tf": &hcl_ast.ObjectList{
					Items: []*hcl_ast.ObjectItem{
						&hcl_ast.ObjectItem{
							Keys: []*hcl_ast.ObjectKey{
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 4,
										Pos: token.Pos{
											Filename: "",
											Offset:   0,
											Line:     1,
											Column:   1,
										},
										Text: "variable",
										JSON: false,
									},
								},
								&hcl_ast.ObjectKey{
									Token: token.Token{
										Type: 9,
										Pos: token.Pos{
											Filename: "",
											Offset:   9,
											Line:     1,
											Column:   10,
										},
										Text: "\"stat\"",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{
								Filename: "",
								Offset:   0,
								Line:     0,
								Column:   0,
							},
							Val: &hcl_ast.ObjectType{
								Lbrace: token.Pos{
									Filename: "",
									Offset:   16,
									Line:     1,
									Column:   17,
								},
								Rbrace: token.Pos{
									Filename: "",
									Offset:   96,
									Line:     6,
									Column:   1,
								},
								List: &hcl_ast.ObjectList{
									Items: []*hcl_ast.ObjectItem{
										&hcl_ast.ObjectItem{
											Keys: []*hcl_ast.ObjectKey{
												&hcl_ast.ObjectKey{
													Token: token.Token{
														Type: 4,
														Pos: token.Pos{
															Filename: "",
															Offset:   22,
															Line:     2,
															Column:   5,
														},
														Text: "default",
														JSON: false,
													},
												},
											},
											Assign: token.Pos{
												Filename: "",
												Offset:   34,
												Line:     2,
												Column:   17,
											},
											Val: &hcl_ast.LiteralType{
												Token: token.Token{
													Type: 9,
													Pos: token.Pos{
														Filename: "",
														Offset:   36,
														Line:     2,
														Column:   19,
													},
													Text: "\"usage\"",
													JSON: false,
												},
												LineComment: (*hcl_ast.CommentGroup)(nil),
											},
											LeadComment: (*hcl_ast.CommentGroup)(nil),
											LineComment: (*hcl_ast.CommentGroup)(nil),
										},
									},
								},
							},
						},
					},
				},
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
		result, err := detectVariables(tc.Input)
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
