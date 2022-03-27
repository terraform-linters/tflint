package tflint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_EvaluateExpr(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		Name     string
		Content  string
		Type     cty.Type
		Want     string
		ErrCheck func(error) bool
	}{
		{
			Name: "string literal",
			Content: `
resource "null_resource" "test" {
  key = "literal_val"
}`,
			Type:     cty.String,
			Want:     `cty.StringVal("literal_val")`,
			ErrCheck: neverHappend,
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
			Type:     cty.String,
			Want:     `cty.StringVal("string_val")`,
			ErrCheck: neverHappend,
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
			Type:     cty.String,
			Want:     `cty.StringVal("string_val")`,
			ErrCheck: neverHappend,
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
			Type:     cty.String,
			Want:     `cty.StringVal("one")`,
			ErrCheck: neverHappend,
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
			Type:     cty.String,
			Want:     `cty.StringVal("one")`,
			ErrCheck: neverHappend,
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
			Type:     cty.String,
			Want:     `cty.StringVal("bar")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "convert to string from integer",
			Content: `
variable "string_var" {
  default = 10
}

resource "null_resource" "test" {
  key = "${var.string_var}"
}`,
			Type:     cty.String,
			Want:     `cty.StringVal("10")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "conditional",
			Content: `
resource "null_resource" "test" {
  key = "${true ? "production" : "development"}"
}`,
			Type:     cty.String,
			Want:     `cty.StringVal("production")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "bulit-in function",
			Content: `
resource "null_resource" "test" {
  key = "${md5("foo")}"
}`,
			Type:     cty.String,
			Want:     `cty.StringVal("acbd18db4cc2f85cedef654fccc4a4d8")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "terraform workspace",
			Content: `
resource "null_resource" "test" {
  key = "${terraform.workspace}"
}`,
			Type:     cty.String,
			Want:     `cty.StringVal("default")`,
			ErrCheck: neverHappend,
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
			Type:     cty.String,
			Want:     `cty.StringVal("Hello World")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "path.root",
			Content: `
resource "null_resource" "test" {
  key = path.root
}`,
			Type:     cty.String,
			Want:     `cty.StringVal(".")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "path.module",
			Content: `
resource "null_resource" "test" {
  key = path.module
}`,
			Type:     cty.String,
			Want:     `cty.StringVal(".")`,
			ErrCheck: neverHappend,
		},
		{
			Name: "integer interpolation",
			Content: `
variable "integer_var" {
  default = 3
}

resource "null_resource" "test" {
  key = "${var.integer_var}"
}`,
			Type:     cty.Number,
			Want:     `cty.NumberIntVal(3)`,
			ErrCheck: neverHappend,
		},
		{
			Name: "convert to integer from string",
			Content: `
variable "integer_var" {
  default = "3"
}

resource "null_resource" "test" {
  key = "${var.integer_var}"
}`,
			Type:     cty.Number,
			Want:     `cty.NumberIntVal(3)`,
			ErrCheck: neverHappend,
		},
		{
			Name: "string list literal",
			Content: `
resource "null_resource" "test" {
  key = ["one", "two", "three"]
}`,
			Type:     cty.List(cty.String),
			Want:     `cty.ListVal([]cty.Value{cty.StringVal("one"), cty.StringVal("two"), cty.StringVal("three")})`,
			ErrCheck: neverHappend,
		},
		{
			Name: "number list literal",
			Content: `
resource "null_resource" "test" {
  key = [1, 2, 3]
}`,
			Type:     cty.List(cty.Number),
			Want:     `cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2), cty.NumberIntVal(3)})`,
			ErrCheck: neverHappend,
		},
		{
			Name: "string map literal",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = "2"
  }
}`,
			Type:     cty.Map(cty.String),
			Want:     `cty.MapVal(map[string]cty.Value{"one":cty.StringVal("1"), "two":cty.StringVal("2")})`,
			ErrCheck: neverHappend,
		},
		{
			Name: "number map literal",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = "2"
  }
}`,
			Type:     cty.Map(cty.Number),
			Want:     `cty.MapVal(map[string]cty.Value{"one":cty.NumberIntVal(1), "two":cty.NumberIntVal(2)})`,
			ErrCheck: neverHappend,
		},
		{
			Name: "map object literal",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = "2"
  }
}`,
			Type:     cty.DynamicPseudoType,
			Want:     `cty.ObjectVal(map[string]cty.Value{"one":cty.NumberIntVal(1), "two":cty.StringVal("2")})`,
			ErrCheck: neverHappend,
		},
		{
			Name: "undefined variable",
			Content: `
resource "null_resource" "test" {
  key = "${var.undefined_var}"
}`,
			Type: cty.String,
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != `failed to eval an expression in main.tf:3; Reference to undeclared input variable: An input variable with the name "undefined_var" has not been declared. This variable can be declared with a variable "undefined_var" {} block.`
			},
		},
		{
			Name: "no default value",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
  key = "${var.no_value_var}"
}`,
			Type: cty.String,
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unknown value found in main.tf:5" || !errors.Is(err, sdk.ErrUnknownValue)
			},
		},
		{
			Name: "no default value as cty.Value",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
  key = "${var.no_value_var}"
}`,
			Type:     cty.DynamicPseudoType,
			Want:     `cty.DynamicVal`,
			ErrCheck: neverHappend,
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
			Type: cty.String,
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "null value found in main.tf:8" || !errors.Is(err, sdk.ErrNullValue)
			},
		},
		{
			Name: "null value as cty.Value",
			Content: `
variable "null_var" {
  type    = string
  default = null
}

resource "null_resource" "test" {
  key = var.null_var
}`,
			Type:     cty.DynamicPseudoType,
			Want:     `cty.NullVal(cty.String)`,
			ErrCheck: neverHappend,
		},
		{
			Name: "terraform env",
			Content: `
resource "null_resource" "test" {
  key = "${terraform.env}"
}`,
			Type: cty.String,
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != `failed to eval an expression in main.tf:3; Invalid "terraform" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The "state environment" concept was renamed to "workspace" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.`
			},
		},
		{
			Name: "type mismatch",
			Content: `
resource "null_resource" "test" {
  key = ["one", "two", "three"]
}`,
			Type: cty.String,
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "failed to eval an expression in main.tf:3; Incorrect value type: Invalid expression value: string required."
			},
		},
		{
			Name: "unevalauble",
			Content: `
resource "null_resource" "test" {
  key = "${module.text}"
}`,
			Type: cty.String,
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unevaluable expression found in main.tf:3" || !errors.Is(err, sdk.ErrUnevaluable)
			},
		},
		{
			Name: "undefined variable in map",
			Content: `
resource "null_resource" "test" {
  key = {
    value = var.undefined_var
  }
}`,
			Type: cty.Map(cty.String),
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != `failed to eval an expression in main.tf:3; Reference to undeclared input variable: An input variable with the name "undefined_var" has not been declared. This variable can be declared with a variable "undefined_var" {} block.`
			},
		},
		{
			Name: "no default value in map",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
  key = {
    value = var.no_value_var
  }
}`,
			Type: cty.Map(cty.String),
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unknown value found in main.tf:5" || !errors.Is(err, sdk.ErrUnknownValue)
			},
		},
		{
			Name: "null value in map",
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
			Type: cty.Map(cty.String),
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "null value found in main.tf:8" || !errors.Is(err, sdk.ErrNullValue)
			},
		},
		{
			Name: "unevalauble in map",
			Content: `
resource "null_resource" "test" {
  key = {
    value = module.text
  }
}`,
			Type: cty.Map(cty.String),
			Want: `cty.NilVal`,
			ErrCheck: func(err error) bool {
				return err == nil || err.Error() != "unevaluable expression found in main.tf:3" || !errors.Is(err, sdk.ErrUnevaluable)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": test.Content})

			body, diags := runner.GetModuleContent(&hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Attributes: []hclext.AttributeSchema{{Name: "key"}},
						},
					},
				},
			}, sdk.GetModuleContentOption{})
			if diags.HasErrors() {
				t.Fatalf("failed to parse: %s", diags)
			}

			resource := body.Blocks[0]
			attribute := resource.Body.Attributes["key"]

			val, err := runner.EvaluateExpr(attribute.Expr, test.Type)
			if test.ErrCheck(err) {
				t.Fatalf("failed to eval: %s", err)
			}

			if test.Want != val.GoString() {
				t.Errorf("`%s` is expected, but got `%s`", test.Want, val.GoString())
			}
		})
	}
}

func Test_EvaluateExpr_pathCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	expected := fmt.Sprintf(`cty.StringVal("%s")`, filepath.ToSlash(cwd))

	content := `
resource "null_resource" "test" {
  key = path.cwd
}`
	runner := TestRunner(t, map[string]string{"main.tf": content})

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "key"}},
				},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		t.Fatalf("failed to parse: %s", diags)
	}

	resource := body.Blocks[0]
	attribute := resource.Body.Attributes["key"]

	val, err := runner.EvaluateExpr(attribute.Expr, cty.String)
	if err != nil {
		t.Fatalf("failed to eval: %s", err)
	}

	if expected != val.GoString() {
		t.Errorf("`%s` is expected, but got `%s`", expected, val.GoString())
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
		t.Run(tc.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

			body, diags := runner.GetModuleContent(&hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Attributes: []hclext.AttributeSchema{{Name: "key"}},
						},
					},
				},
			}, sdk.GetModuleContentOption{})
			if diags.HasErrors() {
				t.Fatalf("failed to parse: %s", diags)
			}

			resource := body.Blocks[0]
			attribute := resource.Body.Attributes["key"]

			ret, err := isEvaluableExpr(attribute.Expr)
			if err != nil && tc.Error == "" {
				t.Fatalf("unexpected error occurred: %s", err)
			}
			if err == nil && tc.Error != "" {
				t.Fatalf("expected error is %s, but no errors", tc.Error)
			}
			if err != nil && tc.Error != "" && err.Error() != tc.Error {
				t.Fatalf("expected error is %s, but got %s", tc.Error, err)
			}
			if ret != tc.Expected {
				t.Fatalf("expected value is %t, but get %t", tc.Expected, ret)
			}
		})
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
			Expected: `cty.StringVal("m4.large")`,
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
				{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("c5.2xlarge"),
						SourceType: terraform.ValueFromNamedFile,
					},
				},
			},
			Expected: `cty.StringVal("c5.2xlarge")`,
		},
		{
			Name: "override variables by variables passed later",
			Content: `
variable "instance_type" {}

resource "null_resource" "test" {
  key = "${var.instance_type}"
}`,
			InputValues: []terraform.InputValues{
				{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("c5.2xlarge"),
						SourceType: terraform.ValueFromNamedFile,
					},
				},
				{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("p3.8xlarge"),
						SourceType: terraform.ValueFromNamedFile,
					},
				},
			},
			Expected: `cty.StringVal("p3.8xlarge")`,
		},
	}

	for _, tc := range cases {
		withEnvVars(t, tc.EnvVar, func() {
			t.Run(tc.Name, func(t *testing.T) {
				runner := testRunnerWithInputVariables(t, map[string]string{"main.tf": tc.Content}, tc.InputValues...)

				body, diags := runner.GetModuleContent(&hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "key"}},
							},
						},
					},
				}, sdk.GetModuleContentOption{})
				if diags.HasErrors() {
					t.Fatalf("failed to parse: %s", diags)
				}

				resource := body.Blocks[0]
				attribute := resource.Body.Attributes["key"]

				val, err := runner.EvaluateExpr(attribute.Expr, cty.String)
				if err != nil {
					t.Fatalf("failed to eval: %s", err)
				}

				if tc.Expected != val.GoString() {
					t.Errorf("%s is expected, but got %s", tc.Expected, val.GoString())
				}
			})
		})
	}
}

func Test_willEvaluateResource(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected int
	}{
		{
			Name: "no meta-arguments",
			Content: `
resource "null_resource" "test" {
}`,
			Expected: 1,
		},
		{
			Name: "count is not zero (literal)",
			Content: `
resource "null_resource" "test" {
  count = 1
}`,
			Expected: 1,
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
			Expected: 1,
		},
		{
			Name: "count is unknown",
			Content: `
variable "foo" {}

resource "null_resource" "test" {
  count = var.foo
}`,
			Expected: 0,
		},
		{
			Name: "count is unevaluable",
			Content: `
resource "null_resource" "test" {
  count = local.foo
}`,
			Expected: 0,
		},
		{
			Name: "count is zero",
			Content: `
resource "null_resource" "test" {
  count = 0
}`,
			Expected: 0,
		},
		{
			// HINT: Terraform does not allow null as `count`
			Name: "count is null",
			Content: `
resource "null_resource" "test" {
  count = null
}`,
			Expected: 1,
		},
		{
			Name: "for_each is not empty (literal)",
			Content: `
resource "null_resource" "test" {
  for_each = {
    foo = "bar"
  }
}`,
			Expected: 1,
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
			Expected: 1,
		},
		{
			Name: "for_each is unknown",
			Content: `
variable "foo" {}

resource "null_resource" "test" {
  for_each = var.foo
}`,
			Expected: 0,
		},
		{
			Name: "for_each is unevaluable",
			Content: `
resource "null_resource" "test" {
  for_each = local.foo
}`,
			Expected: 0,
		},
		{
			Name: "for_each is empty",
			Content: `
resource "null_resource" "test" {
  for_each = {}
}`,
			Expected: 0,
		},
		{
			Name: "for_each is not empty set",
			Content: `
resource "null_resource" "test" {
  for_each = toset(["foo", "bar"])
}`,
			Expected: 1,
		},
		{
			Name: "for_each is empty set",
			Content: `
resource "null_resource" "test" {
  for_each = toset([])
}`,
			Expected: 0,
		},
		{
			// HINT: Terraform does not allow null as `for_each`
			Name: "for_each is null",
			Content: `
resource "null_resource" "test" {
  for_each = null
}`,
			Expected: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

			resources, diags := runner.GetModuleContent(&hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body:       &hclext.BodySchema{},
					},
				},
			}, sdk.GetModuleContentOption{})
			if diags.HasErrors() {
				t.Fatalf("failed to parse: %s", diags)
			}
			if len(resources.Blocks) != tc.Expected {
				t.Fatalf("%d resources expected, but got %d resources", tc.Expected, len(resources.Blocks))
			}
		})
	}
}
