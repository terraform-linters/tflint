package tflint

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_EvaluateExpr_string(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected string
	}{
		{
			Name: "literal",
			Content: `
resource "null_resource" "test" {
  key = "literal_val"
}`,
			Expected: "literal_val",
		},
		{
			Name: "string interpolation",
			Content: `
variable "string_var" {
  default = "string_val"
}

resource "null_resource" "test" {
  key = "${var.string_var}"
}`,
			Expected: "string_val",
		},
		{
			Name: "new style interpolation",
			Content: `
variable "string_var" {
  default = "string_val"
}

resource "null_resource" "test" {
  key = var.string_var
}`,
			Expected: "string_val",
		},
		{
			Name: "list element",
			Content: `
variable "list_var" {
  default = ["one", "two"]
}

resource "null_resource" "test" {
  key = "${var.list_var[0]}"
}`,
			Expected: "one",
		},
		{
			Name: "map element",
			Content: `
variable "map_var" {
  default = {
    one = "one"
    two = "two"
  }
}

resource "null_resource" "test" {
  key = "${var.map_var["one"]}"
}`,
			Expected: "one",
		},
		{
			Name: "object item",
			Content: `
variable "object" {
  type = object({ foo = string })
  default = { foo = "bar" }
}

resource "null_resource" "test" {
  key = var.object.foo
}`,
			Expected: "bar",
		},
		{
			Name: "convert from integer",
			Content: `
variable "string_var" {
  default = 10
}

resource "null_resource" "test" {
  key = "${var.string_var}"
}`,
			Expected: "10",
		},
		{
			Name: "conditional",
			Content: `
resource "null_resource" "test" {
  key = "${true ? "production" : "development"}"
}`,
			Expected: "production",
		},
		{
			Name: "bulit-in function",
			Content: `
resource "null_resource" "test" {
  key = "${md5("foo")}"
}`,
			Expected: "acbd18db4cc2f85cedef654fccc4a4d8",
		},
		{
			Name: "terraform workspace",
			Content: `
resource "null_resource" "test" {
  key = "${terraform.workspace}"
}`,
			Expected: "default",
		},
		{
			Name: "inside interpolation",
			Content: `
variable "string_var" {
  default = "World"
}

resource "null_resource" "test" {
  key = "Hello ${var.string_var}"
}`,
			Expected: "Hello World",
		},
		{
			Name: "path.root",
			Content: `
resource "null_resource" "test" {
  key = path.root
}`,
			Expected: ".",
		},
		{
			Name: "path.module",
			Content: `
resource "null_resource" "test" {
  key = path.module
}`,
			Expected: ".",
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret string
			if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
				return err
			}

			if tc.Expected != ret {
				t.Fatalf("Failed `%s` test: expected value is `%s`, but get `%s`", tc.Name, tc.Expected, ret)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_pathCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.ToSlash(cwd)

	content := `
resource "null_resource" "test" {
  key = path.cwd
}`
	runner := TestRunner(t, map[string]string{"main.tf": content})

	err = runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
		var ret string
		if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
			return err
		}

		if expected != ret {
			t.Fatalf("expected value is `%s`, but get `%s`", expected, ret)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed: `%s` occurred", err)
	}
}

func Test_EvaluateExpr_integer(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected int
	}{
		{
			Name: "integer interpolation",
			Content: `
variable "integer_var" {
  default = 3
}

resource "null_resource" "test" {
  key = "${var.integer_var}"
}`,
			Expected: 3,
		},
		{
			Name: "convert from string",
			Content: `
variable "integer_var" {
  default = "3"
}

resource "null_resource" "test" {
  key = "${var.integer_var}"
}`,
			Expected: 3,
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret int
			if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
				return err
			}

			if tc.Expected != ret {
				t.Fatalf("Failed `%s` test: expected value is `%d`, but get `%d`", tc.Name, tc.Expected, ret)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_stringList(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected []string
	}{
		{
			Name: "list literal",
			Content: `
resource "null_resource" "test" {
  key = ["one", "two", "three"]
}`,
			Expected: []string{"one", "two", "three"},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret []string
			if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
				return err
			}

			if !cmp.Equal(tc.Expected, ret) {
				t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_numberList(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected []int
	}{
		{
			Name: "list literal",
			Content: `
resource "null_resource" "test" {
  key = [1, 2, 3]
}`,
			Expected: []int{1, 2, 3},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret []int
			if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
				return err
			}

			if !cmp.Equal(tc.Expected, ret) {
				t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_stringMap(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected map[string]string
	}{
		{
			Name: "map literal",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = "2"
  }
}`,
			Expected: map[string]string{"one": "1", "two": "2"},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret map[string]string
			if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
				return err
			}

			if !cmp.Equal(tc.Expected, ret) {
				t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_numberMap(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected map[string]int
	}{
		{
			Name: "map literal",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = "2"
  }
}`,
			Expected: map[string]int{"one": 1, "two": 2},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret map[string]int
			if err := runner.EvaluateExpr(attribute.Expr, &ret); err != nil {
				return err
			}

			if !cmp.Equal(tc.Expected, ret) {
				t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_interpolationError(t *testing.T) {
	cases := []struct {
		Name    string
		Content string
		Error   Error
	}{
		{
			Name: "undefined variable",
			Content: `
resource "null_resource" "test" {
  key = "${var.undefined_var}"
}`,
			Error: Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval an expression in main.tf:3; Reference to undeclared input variable: An input variable with the name \"undefined_var\" has not been declared. This variable can be declared with a variable \"undefined_var\" {} block.",
			},
		},
		{
			Name: "no default value",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
  key = "${var.no_value_var}"
}`,
			Error: Error{
				Code:    UnknownValueError,
				Level:   WarningLevel,
				Message: "Unknown value found in main.tf:5; Please use environment variables or tfvars to set the value",
			},
		},
		{
			Name: "null value",
			Content: `
variable "null_var" {
  type    = string
  default = null
}

resource "null_resource" "test" {
  key = var.null_var
}`,
			Error: Error{
				Code:    NullValueError,
				Level:   WarningLevel,
				Message: "Null value found in main.tf:8",
			},
		},
		{
			Name: "terraform env",
			Content: `
resource "null_resource" "test" {
  key = "${terraform.env}"
}`,
			Error: Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval an expression in main.tf:3; Invalid \"terraform\" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The \"state environment\" concept was rename to \"workspace\" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.",
			},
		},
		{
			Name: "type mismatch",
			Content: `
resource "null_resource" "test" {
  key = ["one", "two", "three"]
}`,
			Error: Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval an expression in main.tf:3; Incorrect value type: Invalid expression value: string required.",
			},
		},
		{
			Name: "unevalauble",
			Content: `
resource "null_resource" "test" {
  key = "${module.text}"
}`,
			Error: Error{
				Code:    UnevaluableError,
				Level:   WarningLevel,
				Message: "Unevaluable expression found in main.tf:3",
			},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret string
			err := runner.EvaluateExpr(attribute.Expr, &ret)

			AssertAppError(t, tc.Error, err)
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateExpr_mapWithInterpolationError(t *testing.T) {
	cases := []struct {
		Name    string
		Content string
		Error   Error
	}{
		{
			Name: "undefined variable",
			Content: `
resource "null_resource" "test" {
  key = {
		value = var.undefined_var
	}
}`,
			Error: Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval an expression in main.tf:3; Reference to undeclared input variable: An input variable with the name \"undefined_var\" has not been declared. This variable can be declared with a variable \"undefined_var\" {} block.",
			},
		},
		{
			Name: "no default value",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
	key = {
		value = var.no_value_var
	}
}`,
			Error: Error{
				Code:    UnknownValueError,
				Level:   WarningLevel,
				Message: "Unknown value found in main.tf:5; Please use environment variables or tfvars to set the value",
			},
		},
		{
			Name: "null value",
			Content: `
variable "null_var" {
	type    = string
	default = null
}

resource "null_resource" "test" {
	key = {
		value = var.null_var
	}
}`,
			Error: Error{
				Code:    NullValueError,
				Level:   WarningLevel,
				Message: "Null value found in main.tf:8",
			},
		},
		{
			Name: "unevalauble",
			Content: `
resource "null_resource" "test" {
	key = {
		value = module.text
	}
}`,
			Error: Error{
				Code:    UnevaluableError,
				Level:   WarningLevel,
				Message: "Unevaluable expression found in main.tf:3",
			},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			var ret map[string]string
			err := runner.EvaluateExpr(attribute.Expr, &ret)

			AssertAppError(t, tc.Error, err)
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateBlock(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected map[string]string
	}{
		{
			Name: "map literal",
			Content: `
resource "null_resource" "test" {
  key {
    one = 1
    two = "2"
  }
}`,
			Expected: map[string]string{"one": "1", "two": "2"},
		},
		{
			Name: "variable",
			Content: `
variable "one" {
  default = 1
}

resource "null_resource" "test" {
  key {
    one = var.one
    two = "2"
  }
}`,
			Expected: map[string]string{"one": "1", "two": "2"},
		},
		{
			Name: "null value",
			Content: `
variable "null_var" {
  type    = string
  default = null
}

resource "null_resource" "test" {
  key {
	one = "1"
	two = var.null_var
  }
}`,
			Expected: map[string]string{"one": "1", "two": ""},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceBlocks("null_resource", "key", func(block *hcl.Block) error {
			var ret map[string]string
			schema := &configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"one": {Type: cty.String},
					"two": {Type: cty.String},
				},
			}
			if err := runner.EvaluateBlock(block, schema, &ret); err != nil {
				return err
			}

			if !cmp.Equal(tc.Expected, ret) {
				t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EvaluateBlock_error(t *testing.T) {
	cases := []struct {
		Name    string
		Content string
		Error   Error
	}{
		{
			Name: "undefined variable",
			Content: `
resource "null_resource" "test" {
  key {
	one = "1"
	two = var.undefined_var
  }
}`,
			Error: Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval a block in main.tf:3; Reference to undeclared input variable: An input variable with the name \"undefined_var\" has not been declared. This variable can be declared with a variable \"undefined_var\" {} block.",
			},
		},
		{
			Name: "no default value",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
  key {
    one = "1"
    two = var.no_value_var
  }
}`,
			Error: Error{
				Code:    UnknownValueError,
				Level:   WarningLevel,
				Message: "Unknown value found in main.tf:5; Please use environment variables or tfvars to set the value",
			},
		},
		{
			Name: "type mismatch",
			Content: `
resource "null_resource" "test" {
  key {
    one = "1"
    two = {
      three = 3
    }
  }
}`,
			Error: Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval a block in main.tf:3; Incorrect attribute value type: Inappropriate value for attribute \"two\": string required.",
			},
		},
		{
			Name: "unevalauble",
			Content: `
resource "null_resource" "test" {
  key {
	one = "1"
	two = module.text
  }
}`,
			Error: Error{
				Code:    UnevaluableError,
				Level:   WarningLevel,
				Message: "Unevaluable block found in main.tf:3",
			},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceBlocks("null_resource", "key", func(block *hcl.Block) error {
			var ret map[string]string
			schema := &configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"one": {Type: cty.String},
					"two": {Type: cty.String},
				},
			}
			err := runner.EvaluateBlock(block, schema, &ret)

			AssertAppError(t, tc.Error, err)
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_isEvaluableExpr(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected bool
		Error    string
	}{
		{
			Name: "literal",
			Content: `
resource "null_resource" "test" {
  key = "literal_val"
}`,
			Expected: true,
		},
		{
			Name: "var syntax",
			Content: `
resource "null_resource" "test" {
  key = "${var.string_var}"
}`,
			Expected: true,
		},
		{
			Name: "new var syntax",
			Content: `
resource "null_resource" "test" {
  key = var.string_var
}`,
			Expected: true,
		},
		{
			Name: "conditional",
			Content: `
resource "null_resource" "test" {
  key = "${true ? "production" : "development"}"
}`,
			Expected: true,
		},
		{
			Name: "function",
			Content: `
resource "null_resource" "test" {
  key = "${md5("foo")}"
}`,
			Expected: true,
		},
		{
			Name: "terraform attributes",
			Content: `
resource "null_resource" "test" {
  key = "${terraform.workspace}"
}`,
			Expected: true,
		},
		{
			Name: "include supported syntax",
			Content: `
resource "null_resource" "test" {
  key = "Hello ${var.string_var}"
}`,
			Expected: true,
		},
		{
			Name: "list",
			Content: `
resource "null_resource" "test" {
  key = ["one", "two", "three"]
}`,
			Expected: true,
		},
		{
			Name: "map",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = 2
  }
}`,
			Expected: true,
		},
		{
			Name: "module",
			Content: `
resource "null_resource" "test" {
  key = "${module.text}"
}`,
			Expected: false,
		},
		{
			Name: "resource",
			Content: `
resource "null_resource" "test" {
  key = "${aws_subnet.app.id}"
}`,
			Expected: false,
		},
		{
			Name: "include unsupported syntax",
			Content: `
resource "null_resource" "test" {
  key = "${var.text} ${lookup(var.roles, count.index)}"
}`,
			Expected: false,
		},
		{
			Name: "include unsupported syntax map",
			Content: `
resource "null_resource" "test" {
	key = {
		var = var.text
		unsupported = aws_subnet.app.id
	}
}`,
			Expected: false,
		},
		{
			Name: "path attributes",
			Content: `
resource "null_resource" "test" {
	key = path.cwd
}`,
			Expected: true,
		},
		{
			Name: "invalid reference",
			Content: `
resource "null_resource" "test" {
	key = invalid
}`,
			Expected: false,
			Error:    "Invalid reference: A reference to a resource type must be followed by at least one attribute access, specifying the resource name.",
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
			ret, err := runner.isEvaluableExpr(attribute.Expr)
			if err != nil && tc.Error == "" {
				t.Fatalf("Failed `%s` test: unexpected error occurred: %s", tc.Name, err)
			}
			if err == nil && tc.Error != "" {
				t.Fatalf("Failed `%s` test: expected error is %s, but no errors", tc.Name, tc.Error)
			}
			if err != nil && tc.Error != "" && err.Error() != tc.Error {
				t.Fatalf("Failed `%s` test: expected error is %s, but got %s", tc.Name, tc.Error, err)
			}
			if ret != tc.Expected {
				t.Fatalf("Failed `%s` test: expected value is %t, but get %t", tc.Name, tc.Expected, ret)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_overrideVariables(t *testing.T) {
	cases := []struct {
		Name        string
		Content     string
		EnvVar      map[string]string
		InputValues []terraform.InputValues
		Expected    string
	}{
		{
			Name: "override default value by environment variables",
			Content: `
variable "instance_type" {
  default = "t2.micro"
}

resource "null_resource" "test" {
  key = "${var.instance_type}"
}`,
			EnvVar:   map[string]string{"TF_VAR_instance_type": "m4.large"},
			Expected: "m4.large",
		},
		{
			Name: "override environment variables by passed variables",
			Content: `
variable "instance_type" {}

resource "null_resource" "test" {
  key = "${var.instance_type}"
}`,
			EnvVar: map[string]string{"TF_VAR_instance_type": "m4.large"},
			InputValues: []terraform.InputValues{
				terraform.InputValues{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("c5.2xlarge"),
						SourceType: terraform.ValueFromNamedFile,
					},
				},
			},
			Expected: "c5.2xlarge",
		},
		{
			Name: "override variables by variables passed later",
			Content: `
variable "instance_type" {}

resource "null_resource" "test" {
  key = "${var.instance_type}"
}`,
			InputValues: []terraform.InputValues{
				terraform.InputValues{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("c5.2xlarge"),
						SourceType: terraform.ValueFromNamedFile,
					},
				},
				terraform.InputValues{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("p3.8xlarge"),
						SourceType: terraform.ValueFromNamedFile,
					},
				},
			},
			Expected: "p3.8xlarge",
		},
	}

	for _, tc := range cases {
		withEnvVars(t, tc.EnvVar, func() {
			runner := testRunnerWithInputVariables(t, map[string]string{"main.tf": tc.Content}, tc.InputValues...)

			err := runner.WalkResourceAttributes("null_resource", "key", func(attribute *hcl.Attribute) error {
				var ret string
				err := runner.EvaluateExpr(attribute.Expr, &ret)
				if err != nil {
					return err
				}

				if tc.Expected != ret {
					t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
				}
				return nil
			})

			if err != nil {
				t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
			}
		})
	}
}

func Test_willEvaluateResource(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected bool
	}{
		{
			Name: "no meta-arguments",
			Content: `
resource "null_resource" "test" {
}`,
			Expected: true,
		},
		{
			Name: "count is not zero (literal)",
			Content: `
resource "null_resource" "test" {
  count = 1
}`,
			Expected: true,
		},
		{
			Name: "count is not zero (variable)",
			Content: `
variable "foo" {
  default = 1
}

resource "null_resource" "test" {
  count = var.foo
}`,
			Expected: true,
		},
		{
			Name: "count is unevaluable",
			Content: `
variable "foo" {}

resource "null_resource" "test" {
  count = var.foo
}`,
			Expected: false,
		},
		{
			Name: "count is zero",
			Content: `
resource "null_resource" "test" {
  count = 0
}`,
			Expected: false,
		},
		{
			Name: "for_each is not empty (literal)",
			Content: `
resource "null_resource" "test" {
  for_each = {
    foo = "bar"
  }
}`,
			Expected: true,
		},
		{
			Name: "for_each is not empty (variable)",
			Content: `
variable "object" {
  default = {
    foo = "bar"
  }
}

resource "null_resource" "test" {
  for_each = var.object
}`,
			Expected: true,
		},
		{
			Name: "for_each is unevaluable",
			Content: `
variable "foo" {}

resource "null_resource" "test" {
  for_each = var.foo
}`,
			Expected: false,
		},
		{
			Name: "for_each is empty",
			Content: `
resource "null_resource" "test" {
  for_each = {}
}`,
			Expected: false,
		},
		{
			Name: "for_each is not empty set",
			Content: `
resource "null_resource" "test" {
  for_each = toset(["foo", "bar"])
}`,
			Expected: true,
		},
		{
			Name: "for_each is empty set",
			Content: `
resource "null_resource" "test" {
  for_each = toset([])
}`,
			Expected: false,
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		got, err := runner.willEvaluateResource(runner.LookupResourcesByType("null_resource")[0])
		if err != nil {
			t.Fatalf("Failed `%s`: %s", tc.Name, err)
		}

		if got != tc.Expected {
			t.Fatalf("Failed `%s`: expect to get %t, but got %t", tc.Name, tc.Expected, got)
		}
	}
}

func Test_willEvaluateResource_Error(t *testing.T) {
	cases := []struct {
		Name    string
		Content string
		Error   error
	}{
		{
			Name: "not iterable",
			Content: `
resource "null_resource" "test" {
  for_each = "foo"
}`,
			Error: errors.New("The `for_each` value is not iterable in main.tf:3"),
		},
		{
			Name: "eval error",
			Content: `
resource "null_resource" "test" {
  for_each = var.undefined
}`,
			Error: &Error{
				Code:    EvaluationError,
				Level:   ErrorLevel,
				Message: "Failed to eval an expression in main.tf:3; Reference to undeclared input variable: An input variable with the name \"undefined\" has not been declared. This variable can be declared with a variable \"undefined\" {} block.",
			},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		_, err := runner.willEvaluateResource(runner.LookupResourcesByType("null_resource")[0])
		if err == nil {
			t.Fatalf("Failed `%s`: expected to get an error, but not", tc.Name)
		}
		if err.Error() != tc.Error.Error() {
			t.Fatalf("Failed `%s`: expected to get '%s', but got '%s'", tc.Name, tc.Error.Error(), err.Error())
		}
	}
}
