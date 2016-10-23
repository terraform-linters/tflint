package evaluator

import (
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
)

func TestEval(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Src    string
		Result string
	}{
		{
			Name: "completed string variable",
			Input: `
variable "name" {
    type = "string"
    default = "test"
}`,
			Src:    "${var.name}",
			Result: "test",
		},
		{
			Name: "completed list variable",
			Input: `
variable "name" {
    type = "list"
    default = ["test1", "test2"]
}`,
			Src:    "${var.name[0]}",
			Result: "test1",
		},
		{
			Name: "completed map variable",
			Input: `
variable "name" {
    type = "map"
    default = {
        key = "test1"
        value = "test2"
    }
}`,
			Src:    "${var.name[\"key\"]}",
			Result: "test1",
		},
		{
			Name: "string variable in missing type",
			Input: `
variable "name" {
    default = "test"
}`,
			Src:    "${var.name}",
			Result: "test",
		},
		{
			Name: "list variable in missing key",
			Input: `
variable "name" {
    default = ["test1", "test2"]
}`,
			Src:    "${var.name[0]}",
			Result: "test1",
		},
		{
			Name: "map variable in missing key",
			Input: `
variable "name" {
    default = {
        key = "test1"
        value = "test2"
    }
}`,
			Src:    "${var.name[\"key\"]}",
			Result: "test1",
		},
		{
			Name:   "undefined variable",
			Input:  "",
			Src:    "${var.name}",
			Result: "",
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input))
		list, _ := root.Node.(*ast.ObjectList)
		listmap := map[string]*ast.ObjectList{"testfile": list}

		evaluator, err := NewEvaluator(listmap)
		if err != nil {
			t.Fatalf("Error: %s\n\ntestcase: %s", err, tc.Name)
		}
		result := evaluator.Eval(tc.Src)
		if result != tc.Result {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

func TestEvalError(t *testing.T) {
	cases := []struct {
		Name  string
		Input string
	}{
		{
			Name:  "missing default",
			Input: "variable \"name\" {}",
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input))
		list, _ := root.Node.(*ast.ObjectList)
		listmap := map[string]*ast.ObjectList{"testfile": list}

		_, err := NewEvaluator(listmap)

		if err == nil {
			t.Fatalf("Error: should cause error.\n\ntestcase: %s", tc.Name)
		}
	}
}
