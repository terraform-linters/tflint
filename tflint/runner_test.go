package tflint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/configs"
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
			ret, err := isEvaluableExpr(attribute.Expr)
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

func Test_NewModuleRunners_noModules(t *testing.T) {
	withinFixtureDir(t, "no_modules", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		runners, err := NewModuleRunners(runner)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if len(runners) > 0 {
			t.Fatal("`NewModuleRunners` must not return runners when there is no module")
		}
	})
}

func Test_NewModuleRunners_nestedModules(t *testing.T) {
	withinFixtureDir(t, "nested_modules", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		runners, err := NewModuleRunners(runner)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if len(runners) != 2 {
			t.Fatal("This function must return 2 runners because the config has 2 modules")
		}

		expectedVars := map[string]map[string]*configs.Variable{
			"root": {
				"override": {
					Name:        "override",
					Default:     cty.StringVal("foo"),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module.tf"),
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 20},
					},
				},
				"no_default": {
					Name:        "no_default",
					Default:     cty.StringVal("bar"),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module.tf"),
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				"unknown": {
					Name:        "unknown",
					Default:     cty.UnknownVal(cty.DynamicPseudoType),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module.tf"),
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 19},
					},
				},
			},
			"root.test": {
				"override": {
					Name:        "override",
					Default:     cty.StringVal("foo"),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module1", "resource.tf"),
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 20},
					},
				},
				"no_default": {
					Name:        "no_default",
					Default:     cty.StringVal("bar"),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module1", "resource.tf"),
						Start:    hcl.Pos{Line: 4, Column: 1},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				"unknown": {
					Name:        "unknown",
					Default:     cty.UnknownVal(cty.DynamicPseudoType),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module1", "resource.tf"),
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 19},
					},
				},
			},
		}

		for _, runner := range runners {
			expected, exists := expectedVars[runner.TFConfig.Path.String()]
			if !exists {
				t.Fatalf("`%s` is not found in module runners", runner.TFConfig.Path)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(cty.Type{}, cty.Value{}),
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
			}
			if !cmp.Equal(expected, runner.TFConfig.Module.Variables, opts...) {
				t.Fatalf("`%s` module variables are unmatched: Diff=%s", runner.TFConfig.Path, cmp.Diff(expected, runner.TFConfig.Module.Variables, opts...))
			}
		}
	})
}

func Test_NewModuleRunners_modVars(t *testing.T) {
	withinFixtureDir(t, "nested_module_vars", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		runners, err := NewModuleRunners(runner)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if len(runners) != 2 {
			t.Fatal("This function must return 2 runners because the config has 2 modules")
		}

		child := runners[0]
		if child.TFConfig.Path.String() != "module1" {
			t.Fatalf("Expected child config path name is `module1`, but get `%s`", child.TFConfig.Path.String())
		}

		expected := map[string]*moduleVariable{
			"foo": {
				Root: true,
				DeclRange: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 4, Column: 9},
					End:      hcl.Pos{Line: 4, Column: 14},
				},
			},
			"bar": {
				Root: true,
				DeclRange: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 5, Column: 9},
					End:      hcl.Pos{Line: 5, Column: 14},
				},
			},
		}
		opts := []cmp.Option{cmpopts.IgnoreFields(hcl.Pos{}, "Byte")}
		if !cmp.Equal(expected, child.modVars, opts...) {
			t.Fatalf("`%s` module variables are unmatched: Diff=%s", child.TFConfig.Path.String(), cmp.Diff(expected, child.modVars, opts...))
		}

		grandchild := runners[1]
		if grandchild.TFConfig.Path.String() != "module1.module2" {
			t.Fatalf("Expected child config path name is `module1.module2`, but get `%s`", grandchild.TFConfig.Path.String())
		}

		expected = map[string]*moduleVariable{
			"red": {
				Root:    false,
				Parents: []*moduleVariable{expected["foo"], expected["bar"]},
				DeclRange: hcl.Range{
					Filename: filepath.Join("module", "main.tf"),
					Start:    hcl.Pos{Line: 8, Column: 11},
					End:      hcl.Pos{Line: 8, Column: 34},
				},
			},
			"blue": {
				Root:    false,
				Parents: []*moduleVariable{},
				DeclRange: hcl.Range{
					Filename: filepath.Join("module", "main.tf"),
					Start:    hcl.Pos{Line: 9, Column: 11},
					End:      hcl.Pos{Line: 9, Column: 17},
				},
			},
			"green": {
				Root:    false,
				Parents: []*moduleVariable{expected["foo"]},
				DeclRange: hcl.Range{
					Filename: filepath.Join("module", "main.tf"),
					Start:    hcl.Pos{Line: 10, Column: 11},
					End:      hcl.Pos{Line: 10, Column: 49},
				},
			},
		}
		opts = []cmp.Option{cmpopts.IgnoreFields(hcl.Pos{}, "Byte")}
		if !cmp.Equal(expected, grandchild.modVars, opts...) {
			t.Fatalf("`%s` module variables are unmatched: Diff=%s", grandchild.TFConfig.Path.String(), cmp.Diff(expected, grandchild.modVars, opts...))
		}
	})
}

func Test_NewModuleRunners_ignoreModules(t *testing.T) {
	withinFixtureDir(t, "nested_modules", func() {
		config := moduleConfig()
		config.IgnoreModules["./module"] = true
		runner := testRunnerWithOsFs(t, config)

		runners, err := NewModuleRunners(runner)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if len(runners) != 0 {
			t.Fatalf("This function must not return runners because `ignore_module` is set. Got `%d` runner(s)", len(runners))
		}
	})
}

func Test_NewModuleRunners_withInvalidExpression(t *testing.T) {
	withinFixtureDir(t, "invalid_module_attribute", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		_, err := NewModuleRunners(runner)

		expected := Error{
			Code:    EvaluationError,
			Level:   ErrorLevel,
			Message: "Failed to eval an expression in module.tf:4; Invalid \"terraform\" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The \"state environment\" concept was rename to \"workspace\" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.",
		}
		AssertAppError(t, expected, err)
	})
}

func Test_NewModuleRunners_withNotAllowedAttributes(t *testing.T) {
	withinFixtureDir(t, "not_allowed_module_attribute", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		_, err := NewModuleRunners(runner)

		expected := Error{
			Code:    UnexpectedAttributeError,
			Level:   ErrorLevel,
			Message: "Attribute of module not allowed was found in module.tf:1; module.tf:4,3-10: Unexpected \"invalid\" block; Blocks are not allowed here.",
		}
		AssertAppError(t, expected, err)
	})
}

func Test_LookupResourcesByType(t *testing.T) {
	content := `
resource "aws_instance" "web" {
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}

resource "aws_route53_zone" "primary" {
  name = "example.com"
}

resource "aws_route" "r" {
  route_table_id            = "rtb-4fbb3ac4"
  destination_cidr_block    = "10.0.1.0/22"
  vpc_peering_connection_id = "pcx-45ff3dc1"
  depends_on                = ["aws_route_table.testing"]
}`

	runner := TestRunner(t, map[string]string{"resource.tf": content})
	resources := runner.LookupResourcesByType("aws_instance")

	if len(resources) != 1 {
		t.Fatalf("Expected resources size is `1`, but get `%d`", len(resources))
	}
	if resources[0].Type != "aws_instance" {
		t.Fatalf("Expected resource type is `aws_instance`, but get `%s`", resources[0].Type)
	}
}

func Test_LookupIssues(t *testing.T) {
	runner := TestRunner(t, map[string]string{})
	runner.Issues = Issues{
		{
			Rule:    &testRule{},
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "template.tf",
				Start:    hcl.Pos{Line: 1},
			},
		},
		{
			Rule:    &testRule{},
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "resource.tf",
				Start:    hcl.Pos{Line: 1},
			},
		},
	}

	ret := runner.LookupIssues("template.tf")
	expected := Issues{
		{
			Rule:    &testRule{},
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "template.tf",
				Start:    hcl.Pos{Line: 1},
			},
		},
	}

	if !cmp.Equal(expected, ret) {
		t.Fatalf("Failed test: diff: %s", cmp.Diff(expected, ret))
	}
}

func Test_WalkResourceAttributes(t *testing.T) {
	cases := []struct {
		Name      string
		Content   string
		ErrorText string
	}{
		{
			Name: "Resource not found",
			Content: `
resource "null_resource" "test" {
  key = "foo"
}`,
		},
		{
			Name: "Attribute not found",
			Content: `
resource "aws_instance" "test" {
  key = "foo"
}`,
		},
		{
			Name: "Block attribute",
			Content: `
resource "aws_instance" "test" {
  instance_type {
    name = "t2.micro"
  }
}`,
		},
		{
			Name: "walk",
			Content: `
resource "aws_instance" "test" {
  instance_type = "t2.micro"
}`,
			ErrorText: "Walk instance_type",
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
			return fmt.Errorf("Walk %s", attribute.Name)
		})
		if err == nil {
			if tc.ErrorText != "" {
				t.Fatalf("Failed `%s` test: expected error is not occurred `%s`", tc.Name, tc.ErrorText)
			}
		} else if err.Error() != tc.ErrorText {
			t.Fatalf("Failed `%s` test: expected error is %s, but get %s", tc.Name, tc.ErrorText, err)
		}
	}
}

func Test_WalkResourceBlocks(t *testing.T) {
	cases := []struct {
		Name      string
		Content   string
		ErrorText string
	}{
		{
			Name: "Resource not found",
			Content: `
resource "null_resource" "test" {
  key {
    foo = "bar"
  }
}`,
		},
		{
			Name: "Block not found",
			Content: `
resource "aws_instance" "test" {
  key {
    foo = "bar"
  }
}`,
		},
		{
			Name: "Attribute",
			Content: `
resource "aws_instance" "test" {
  instance_type = "foo"
}`,
		},
		{
			Name: "walk",
			Content: `
resource "aws_instance" "test" {
  instance_type {
    foo = "bar"
  }
}`,
			ErrorText: "Walk instance_type",
		},
		{
			Name: "walk dynamic blocks",
			Content: `
resource "aws_instance" "test" {
  dynamic "instance_type" {
    for_each = ["foo", "bar"]

    content {
      foo = instance_type.value
    }
  }
}`,
			ErrorText: "Walk content",
		},
		{
			Name: "Another dynamic block",
			Content: `
resource "aws_instance" "test" {
  dynamic "key" {
    for_each = ["foo", "bar"]

    content {
      foo = key.value
    }
  }
}`,
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		err := runner.WalkResourceBlocks("aws_instance", "instance_type", func(block *hcl.Block) error {
			return fmt.Errorf("Walk %s", block.Type)
		})
		if err == nil {
			if tc.ErrorText != "" {
				t.Fatalf("Failed `%s` test: expected error is not occurred `%s`", tc.Name, tc.ErrorText)
			}
		} else if err.Error() != tc.ErrorText {
			t.Fatalf("Failed `%s` test: expected error is %s, but get %s", tc.Name, tc.ErrorText, err)
		}
	}
}

func Test_EnsureNoError(t *testing.T) {
	cases := []struct {
		Name      string
		Error     error
		ErrorText string
	}{
		{
			Name:      "no error",
			Error:     nil,
			ErrorText: "function called",
		},
		{
			Name:      "native error",
			Error:     errors.New("Error occurred"),
			ErrorText: "Error occurred",
		},
		{
			Name: "warning error",
			Error: &Error{
				Code:    UnknownValueError,
				Level:   WarningLevel,
				Message: "Warning error",
			},
		},
		{
			Name: "app error",
			Error: &Error{
				Code:    TypeMismatchError,
				Level:   ErrorLevel,
				Message: "App error",
			},
			ErrorText: "App error",
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{})

		err := runner.EnsureNoError(tc.Error, func() error {
			return errors.New("function called")
		})
		if err == nil {
			if tc.ErrorText != "" {
				t.Fatalf("Failed `%s` test: expected error is not occurred `%s`", tc.Name, tc.ErrorText)
			}
		} else if err.Error() != tc.ErrorText {
			t.Fatalf("Failed `%s` test: expected error is %s, but get %s", tc.Name, tc.ErrorText, err)
		}
	}
}

func Test_IsNullExpr(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected bool
		Error    string
	}{
		{
			Name: "non null literal",
			Content: `
resource "null_resource" "test" {
  key = "string"
}`,
			Expected: false,
		},
		{
			Name: "non null variable",
			Content: `
variable "value" {
  default = "string"
}

resource "null_resource" "test" {
  key = var.value
}`,
			Expected: false,
		},
		{
			Name: "null literal",
			Content: `
resource "null_resource" "test" {
  key = null
}`,
			Expected: true,
		},
		{
			Name: "null variable",
			Content: `
variable "value" {
  default = null
}
	
resource "null_resource" "test" {
  key = var.value
}`,
			Expected: true,
		},
		{
			Name: "unknown variable",
			Content: `
variable "value" {}
	
resource "null_resource" "test" {
  key = var.value
}`,
			Expected: false,
		},
		{
			Name: "unevaluable reference",
			Content: `
resource "null_resource" "test" {
  key = aws_instance.id
}`,
			Expected: false,
		},
		{
			Name: "including null literal",
			Content: `
resource "null_resource" "test" {
  key = "${null}-1"
}`,
			Expected: false,
			Error:    "Invalid template interpolation value: The expression result is null. Cannot include a null value in a string template.",
		},
		{
			Name: "invalid references",
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
			ret, err := runner.IsNullExpr(attribute.Expr)
			if err != nil && tc.Error == "" {
				t.Fatalf("Failed `%s` test: unexpected error occurred: %s", tc.Name, err)
			}
			if err == nil && tc.Error != "" {
				t.Fatalf("Failed `%s` test: expected error is %s, but no errors", tc.Name, tc.Error)
			}
			if err != nil && tc.Error != "" && err.Error() != tc.Error {
				t.Fatalf("Failed `%s` test: expected error is %s, but got %s", tc.Name, tc.Error, err)
			}
			if tc.Expected != ret {
				t.Fatalf("Failed `%s` test: expected value is %t, but get %t", tc.Name, tc.Expected, ret)
			}
			return nil
		})

		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}
	}
}

func Test_EachStringSliceExprs(t *testing.T) {
	cases := []struct {
		Name    string
		Content string
		Vals    []string
		Lines   []int
	}{
		{
			Name: "literal list",
			Content: `
resource "null_resource" "test" {
  value = [
    "text",
    "element",
  ]
}`,
			Vals:  []string{"text", "element"},
			Lines: []int{4, 5},
		},
		{
			Name: "literal list",
			Content: `
variable "list" {
  default = [
    "text",
    "element",
  ]
}

resource "null_resource" "test" {
  value = var.list
}`,
			Vals:  []string{"text", "element"},
			Lines: []int{10, 10},
		},
		{
			Name: "for expressions",
			Content: `
variable "list" {
  default = ["text", "element", "ignored"]
}

resource "null_resource" "test" {
  value = [
	for e in var.list:
	e
	if e != "ignored"
  ]
}`,
			Vals:  []string{"text", "element"},
			Lines: []int{7, 7},
		},
	}

	for _, tc := range cases {
		runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

		vals := []string{}
		lines := []int{}
		err := runner.WalkResourceAttributes("null_resource", "value", func(attribute *hcl.Attribute) error {
			return runner.EachStringSliceExprs(attribute.Expr, func(val string, expr hcl.Expression) {
				vals = append(vals, val)
				lines = append(lines, expr.Range().Start.Line)
			})
		})
		if err != nil {
			t.Fatalf("Failed `%s` test: %s", tc.Name, err)
		}

		if !cmp.Equal(vals, tc.Vals) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(vals, tc.Vals))
		}
		if !cmp.Equal(lines, tc.Lines) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(lines, tc.Lines))
		}
	}
}

type testRule struct{}

func (r *testRule) Name() string {
	return "test_rule"
}
func (r *testRule) Severity() string {
	return ERROR
}
func (r *testRule) Link() string {
	return ""
}

func Test_EmitIssue(t *testing.T) {
	cases := []struct {
		Name        string
		Rule        Rule
		Message     string
		Location    hcl.Range
		Annotations map[string]Annotations
		Expected    Issues
	}{
		{
			Name:    "basic",
			Rule:    &testRule{},
			Message: "This is test message",
			Location: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1},
			},
			Annotations: map[string]Annotations{},
			Expected: Issues{
				{
					Rule:    &testRule{},
					Message: "This is test message",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1},
					},
				},
			},
		},
		{
			Name:    "ignore",
			Rule:    &testRule{},
			Message: "This is test message",
			Location: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1},
			},
			Annotations: map[string]Annotations{
				"test.tf": {
					{
						Content: "test_rule",
						Token: hclsyntax.Token{
							Type: hclsyntax.TokenComment,
							Range: hcl.Range{
								Filename: "test.tf",
								Start:    hcl.Pos{Line: 1},
							},
						},
					},
				},
			},
			Expected: Issues{},
		},
	}

	for _, tc := range cases {
		runner := testRunnerWithAnnotations(t, map[string]string{}, tc.Annotations)

		runner.EmitIssue(tc.Rule, tc.Message, tc.Location)

		if !cmp.Equal(runner.Issues, tc.Expected) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(runner.Issues, tc.Expected))
		}
	}
}
