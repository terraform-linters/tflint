package detector

import (
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/config"
	eval "github.com/wata727/tflint/evaluator"
)

// TODO: add Detect test
//       add (d *Detector) detect test

func TestHclLiteralToken(t *testing.T) {
	type Input struct {
		File string
		Key  string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result token.Token
		Error  bool
	}{
		{
			Name: "return literal token",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "instance_type",
			},
			Result: token.Token{
				Type: 9,
				Pos: token.Pos{
					Filename: "",
					Offset:   47,
					Line:     3,
					Column:   21,
				},
				Text: "\"t2.micro\"",
				JSON: false,
			},
			Error: false,
		},
		{
			Name: "happen error when value is list",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = ["t2.micro"]
}`,
				Key: "instance_type",
			},
			Result: token.Token{},
			Error:  true,
		},
		{
			Name: "happen error when value is map",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = {
        default = "t2.micro"
    }
}`,
				Key: "instance_type",
			},
			Result: token.Token{},
			Error:  true,
		},
		{
			Name: "happen error when key not found",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "ami_id",
			},
			Result: token.Token{},
			Error:  true,
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]

		result, err := hclLiteralToken(item, tc.Input.Key)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if result.Text != tc.Result.Text {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

func TestIsKeyNotFound(t *testing.T) {
	type Input struct {
		File string
		Key  string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result bool
	}{
		{
			Name: "key found",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "instance_type",
			},
			Result: false,
		},
		{
			Name: "happen error when value is list",
			Input: Input{
				File: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
}`,
				Key: "iam_instance_profile",
			},
			Result: true,
		},
	}

	for _, tc := range cases {
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		item := list.Filter("resource", "aws_instance").Items[0]
		result := IsKeyNotFound(item, tc.Input.Key)

		if result != tc.Result {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

func TestEvalToString(t *testing.T) {
	type Input struct {
		Src  string
		File string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result string
		Error  bool
	}{
		{
			Name: "return string",
			Input: Input{
				Src: "${var.text}",
				File: `
variable "text" {
    default = "result"
}`,
			},
			Result: "result",
			Error:  false,
		},
		{
			Name: "not string",
			Input: Input{
				Src: "${var.text}",
				File: `
variable "text" {
    default = ["result"]
}`,
			},
			Result: "",
			Error:  true,
		},
		{
			Name: "not evaluable",
			Input: Input{
				Src:  "${aws_instance.app}",
				File: `variable "text" {}`,
			},
			Result: "",
			Error:  true,
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Input.File))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["text.tf"] = list

		evalConfig, _ := eval.NewEvaluator(listMap, config.Init())
		d := &Detector{
			ListMap:    listMap,
			EvalConfig: evalConfig,
		}

		result, err := d.evalToString(tc.Input.Src)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if result != tc.Result {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}
