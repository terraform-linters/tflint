package tflint

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	}

	dir, err := ioutil.TempDir("", "string")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, map[string]*terraform.InputValue{})

		var ret string
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if tc.Expected != ret {
			t.Fatalf("Failed `%s` test: expected value is `%s`, but get `%s`", tc.Name, tc.Expected, ret)
		}
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

	dir, err := ioutil.TempDir("", "integer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, map[string]*terraform.InputValue{})

		var ret int
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if tc.Expected != ret {
			t.Fatalf("Failed `%s` test: expected value is %d, but get %d", tc.Name, tc.Expected, ret)
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

	dir, err := ioutil.TempDir("", "stringList")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, map[string]*terraform.InputValue{})

		ret := []string{}
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
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

	dir, err := ioutil.TempDir("", "numberList")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, map[string]*terraform.InputValue{})

		ret := []int{}
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}

func Test_EvaluateExpr_map(t *testing.T) {
	type mapObject struct {
		One int `cty:"one"`
		Two int `cty:"two"`
	}

	cases := []struct {
		Name     string
		Content  string
		Expected mapObject
	}{
		{
			Name: "map literal",
			Content: `
resource "null_resource" "test" {
  key = {
    one = 1
    two = 2
  }
}`,
			Expected: mapObject{One: 1, Two: 2},
		},
	}

	dir, err := ioutil.TempDir("", "map")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, map[string]*terraform.InputValue{})

		ret := mapObject{}
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}

func Test_EvaluateExpr_interpolationError(t *testing.T) {
	cases := []struct {
		Name       string
		Content    string
		ErrorCode  int
		ErrorLevel int
		ErrorText  string
	}{
		{
			Name: "undefined variable",
			Content: `
resource "null_resource" "test" {
  key = "${var.undefined_var}"
}`,
			ErrorCode:  EvaluationError,
			ErrorLevel: ErrorLevel,
			ErrorText:  "Failed to eval an expression in resource.tf:3; Reference to undeclared input variable: An input variable with the name \"undefined_var\" has not been declared. This variable can be declared with a variable \"undefined_var\" {} block.",
		},
		{
			Name: "no default value",
			Content: `
variable "no_value_var" {}

resource "null_resource" "test" {
  key = "${var.no_value_var}"
}`,
			ErrorCode:  UnknownValueError,
			ErrorLevel: WarningLevel,
			ErrorText:  "Unknown value found in resource.tf:5; Please use environment variables or tfvars to set the value",
		},
		{
			Name: "terraform env",
			Content: `
resource "null_resource" "test" {
  key = "${terraform.env}"
}`,
			ErrorCode:  EvaluationError,
			ErrorLevel: ErrorLevel,
			ErrorText:  "Failed to eval an expression in resource.tf:3; Invalid \"terraform\" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The \"state environment\" concept was rename to \"workspace\" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.",
		},
		{
			Name: "type mismatch",
			Content: `
resource "null_resource" "test" {
  key = ["one", "two", "three"]
}`,
			ErrorCode:  TypeConversionError,
			ErrorLevel: ErrorLevel,
			ErrorText:  "Invalid type expression in resource.tf:3; incorrect type; string required",
		},
		{
			Name: "unevalauble",
			Content: `
resource "null_resource" "test" {
  key = "${module.text}"
}`,
			ErrorCode:  UnevaluableError,
			ErrorLevel: WarningLevel,
			ErrorText:  "Unevaluable expression found in resource.tf:3",
		},
	}

	dir, err := ioutil.TempDir("", "interpolationError")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, map[string]*terraform.InputValue{})

		var ret string
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if appErr, ok := err.(*Error); ok {
			if appErr == nil {
				t.Fatalf("Failed `%s` test: expected err is `%s`, but nothing occurred", tc.Name, tc.ErrorText)
			}
			if appErr.Code != tc.ErrorCode {
				t.Fatalf("Failed `%s` test: expected error code is `%d`, but get `%d`", tc.Name, tc.ErrorCode, appErr.Code)
			}
			if appErr.Level != tc.ErrorLevel {
				t.Fatalf("Failed `%s` test: expected error level is `%d`, but get `%d`", tc.Name, tc.ErrorLevel, appErr.Level)
			}
			if appErr.Error() != tc.ErrorText {
				t.Fatalf("Failed `%s` test: expected error is `%s`, but get `%s`", tc.Name, tc.ErrorText, appErr.Error())
			}
		} else {
			t.Fatalf("Failed `%s` test: unexpected error occurred: %s", tc.Name, err)
		}
	}
}

func Test_isEvaluable(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected bool
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
	}

	dir, err := ioutil.TempDir("", "isEvaluable")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		ret := isEvaluable(attribute.Expr)
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %t, but get %t", tc.Name, tc.Expected, ret)
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
						SourceType: terraform.ValueFromFile,
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
						SourceType: terraform.ValueFromFile,
					},
				},
				terraform.InputValues{
					"instance_type": &terraform.InputValue{
						Value:      cty.StringVal("p3.8xlarge"),
						SourceType: terraform.ValueFromFile,
					},
				},
			},
			Expected: "p3.8xlarge",
		},
	}

	dir, err := ioutil.TempDir("", "overrideVariables")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		err := ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := loadConfigHelper(dir)
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper("key", cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := NewRunner(cfg, tc.InputValues...)

		var ret string
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if tc.Expected != ret {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
