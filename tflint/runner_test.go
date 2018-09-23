package tflint

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/terraform"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/issue"
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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

		ret := map[string]string{}
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if err != nil {
			t.Fatalf("Failed `%s` test: `%s` occurred", tc.Name, err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff: %s", tc.Name, cmp.Diff(tc.Expected, ret))
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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

		ret := map[string]int{}
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
			ErrorText:  "Invalid type expression in resource.tf:3; string required",
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

		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

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

		runner := NewRunner(EmptyConfig(), cfg, tc.InputValues...)

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

func Test_NewModuleRunners_noModules(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "no_modules"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	cfg, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})
	runners, err := NewModuleRunners(runner)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if len(runners) > 0 {
		t.Fatal("`NewModuleRunners` must not return runners when there is no module")
	}
}

func Test_NewModuleRunners_nestedModules(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "nested_modules"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	cfg, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})
	runners, err := NewModuleRunners(runner)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if len(runners) != 2 {
		t.Fatal("This function must return 2 runners because the config has 2 modules")
	}

	child := runners[0].TFConfig
	if child.Path.String() != "root" {
		t.Fatalf("Expected child config path name is `root`, but get `%s`", child.Path.String())
	}

	expected := map[string]*configs.Variable{
		"override": {
			Name:        "override",
			Default:     cty.StringVal("foo"),
			Type:        cty.DynamicPseudoType,
			ParsingMode: configs.VariableParseLiteral,
			DeclRange: hcl.Range{
				Filename: filepath.Join(".terraform", "modules", "07be448a6067a2bba065bff4beea229d", "module.tf"),
				Start: hcl.Pos{
					Line:   1,
					Column: 1,
					Byte:   0,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 20,
					Byte:   19,
				},
			},
		},
		"no_default": {
			Name:        "no_default",
			Default:     cty.StringVal("bar"),
			Type:        cty.DynamicPseudoType,
			ParsingMode: configs.VariableParseLiteral,
			DeclRange: hcl.Range{
				Filename: filepath.Join(".terraform", "modules", "07be448a6067a2bba065bff4beea229d", "module.tf"),
				Start: hcl.Pos{
					Line:   4,
					Column: 1,
					Byte:   42,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 22,
					Byte:   63,
				},
			},
		},
		"unknown": {
			Name:        "unknown",
			Default:     cty.UnknownVal(cty.DynamicPseudoType),
			Type:        cty.DynamicPseudoType,
			ParsingMode: configs.VariableParseLiteral,
			DeclRange: hcl.Range{
				Filename: filepath.Join(".terraform", "modules", "07be448a6067a2bba065bff4beea229d", "module.tf"),
				Start: hcl.Pos{
					Line:   5,
					Column: 1,
					Byte:   67,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 19,
					Byte:   85,
				},
			},
		},
	}
	if !reflect.DeepEqual(expected, child.Module.Variables) {
		t.Fatalf("`%s` module variables are unmatch:\n Expected: %s\n Actual: %s", child.Path.String(), pp.Sprint(expected), pp.Sprint(child.Module.Variables))
	}

	grandchild := runners[1].TFConfig
	if grandchild.Path.String() != "root.test" {
		t.Fatalf("Expected child config path name is `root.test`, but get `%s`", grandchild.Path.String())
	}

	expected = map[string]*configs.Variable{
		"override": {
			Name:        "override",
			Default:     cty.StringVal("foo"),
			Type:        cty.DynamicPseudoType,
			ParsingMode: configs.VariableParseLiteral,
			DeclRange: hcl.Range{
				Filename: filepath.Join(".terraform", "modules", "a8d8930bc3c2ae53bf6e3bbcb3083d7b", "resource.tf"),
				Start: hcl.Pos{
					Line:   1,
					Column: 1,
					Byte:   0,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 20,
					Byte:   19,
				},
			},
		},
		"no_default": {
			Name:        "no_default",
			Default:     cty.StringVal("bar"),
			Type:        cty.DynamicPseudoType,
			ParsingMode: configs.VariableParseLiteral,
			DeclRange: hcl.Range{
				Filename: filepath.Join(".terraform", "modules", "a8d8930bc3c2ae53bf6e3bbcb3083d7b", "resource.tf"),
				Start: hcl.Pos{
					Line:   4,
					Column: 1,
					Byte:   42,
				},
				End: hcl.Pos{
					Line:   4,
					Column: 22,
					Byte:   63,
				},
			},
		},
		"unknown": {
			Name:        "unknown",
			Default:     cty.UnknownVal(cty.DynamicPseudoType),
			Type:        cty.DynamicPseudoType,
			ParsingMode: configs.VariableParseLiteral,
			DeclRange: hcl.Range{
				Filename: filepath.Join(".terraform", "modules", "a8d8930bc3c2ae53bf6e3bbcb3083d7b", "resource.tf"),
				Start: hcl.Pos{
					Line:   5,
					Column: 1,
					Byte:   67,
				},
				End: hcl.Pos{
					Line:   5,
					Column: 19,
					Byte:   85,
				},
			},
		},
	}
	if !reflect.DeepEqual(expected, grandchild.Module.Variables) {
		t.Fatalf("`%s` module variables are unmatch:\n Expected: %s\n Actual: %s", child.Path.String(), pp.Sprint(expected), pp.Sprint(grandchild.Module.Variables))
	}
}

func Test_NewModuleRunners_ignoreModules(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "nested_modules"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	cfg, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	conf := EmptyConfig()
	conf.IgnoreModule["./module"] = true

	runner := NewRunner(conf, cfg, map[string]*terraform.InputValue{})
	runners, err := NewModuleRunners(runner)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if len(runners) != 0 {
		t.Fatalf("This function must not return runners because `ignore_module` is set. Got `%d` runner(s)", len(runners))
	}
}

func Test_NewModuleRunners_withInvalidExpression(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "invalid_module_attribute"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	cfg, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})
	_, err = NewModuleRunners(runner)

	errText := "Failed to eval an expression in module.tf:4; Invalid \"terraform\" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The \"state environment\" concept was rename to \"workspace\" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute."
	errCode := EvaluationError
	errLevel := ErrorLevel

	if appErr, ok := err.(*Error); ok {
		if appErr == nil {
			t.Fatalf("Expected err is `%s`, but nothing occurred", errText)
		}
		if appErr.Code != errCode {
			t.Fatalf("Expected error code is `%d`, but get `%d`", errCode, appErr.Code)
		}
		if appErr.Level != errLevel {
			t.Fatalf("Expected error level is `%d`, but get `%d`", errLevel, appErr.Level)
		}
		if appErr.Error() != errText {
			t.Fatalf("Expected error is `%s`, but get `%s`", errText, appErr.Error())
		}
	} else {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
}

func Test_NewModuleRunners_withNotAllowedAttributes(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "not_allowed_module_attribute"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	cfg, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})
	_, err = NewModuleRunners(runner)

	errText := "Attribute of module not allowed was found in module.tf:1; Unexpected invalid block; Blocks are not allowed here."
	errCode := UnexpectedAttributeError
	errLevel := ErrorLevel

	if appErr, ok := err.(*Error); ok {
		if appErr == nil {
			t.Fatalf("Expected err is `%s`, but nothing occurred", errText)
		}
		if appErr.Code != errCode {
			t.Fatalf("Expected error code is `%d`, but get `%d`", errCode, appErr.Code)
		}
		if appErr.Level != errLevel {
			t.Fatalf("Expected error level is `%d`, but get `%d`", errLevel, appErr.Level)
		}
		if appErr.Error() != errText {
			t.Fatalf("Expected error is `%s`, but get `%s`", errText, appErr.Error())
		}
	} else {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
}

func Test_LookupResourcesByType(t *testing.T) {
	dir, err := ioutil.TempDir("", "lookupResourcesByType")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = ioutil.WriteFile(dir+"/resource.tf", []byte(`
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
}
`), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfigHelper(dir)
	if err != nil {
		t.Fatal(err)
	}

	runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})
	resources := runner.LookupResourcesByType("aws_instance")

	if len(resources) != 1 {
		t.Fatalf("Expected resources size is `1`, but get `%d`", len(resources))
	}
	if resources[0].Type != "aws_instance" {
		t.Fatalf("Expected resource type is `aws_instance`, but get `%s`", resources[0].Type)
	}
}

func Test_LookupIssues(t *testing.T) {
	runner := NewRunner(EmptyConfig(), configs.NewEmptyConfig(), map[string]*terraform.InputValue{})
	runner.Issues = issue.Issues{
		{
			Detector: "test rule",
			Type:     issue.ERROR,
			Message:  "This is test rule",
			Line:     1,
			File:     "template.tf",
		},
		{
			Detector: "test rule",
			Type:     issue.ERROR,
			Message:  "This is test rule",
			Line:     1,
			File:     "resource.tf",
		},
	}

	ret := runner.LookupIssues("template.tf")
	expected := issue.Issues{
		{
			Detector: "test rule",
			Type:     issue.ERROR,
			Message:  "This is test rule",
			Line:     1,
			File:     "template.tf",
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

	dir, err := ioutil.TempDir("", "WalkResourceAttributes")
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
		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

		err = runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
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

	dir, err := ioutil.TempDir("", "EnsureNoError")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		runner := NewRunner(EmptyConfig(), configs.NewEmptyConfig(), map[string]*terraform.InputValue{})

		err = runner.EnsureNoError(tc.Error, func() error {
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
	}

	dir, err := ioutil.TempDir("", "EachStringSliceExprs")
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
		runner := NewRunner(EmptyConfig(), cfg, map[string]*terraform.InputValue{})

		vals := []string{}
		lines := []int{}
		err = runner.WalkResourceAttributes("null_resource", "value", func(attribute *hcl.Attribute) error {
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
func (r *testRule) Type() string {
	return issue.ERROR
}
func (r *testRule) Link() string {
	return ""
}

func Test_EmitIssue(t *testing.T) {
	cases := []struct {
		Name     string
		Rule     Rule
		Message  string
		Location hcl.Range
		Expected issue.Issues
	}{
		{
			Name:    "basic",
			Rule:    &testRule{},
			Message: "This is test message",
			Location: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1},
			},
			Expected: issue.Issues{
				{
					Detector: "test_rule",
					Type:     issue.ERROR,
					Message:  "This is test message",
					Line:     1,
					File:     "test.tf",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "EmitIssue")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		runner := NewRunner(EmptyConfig(), configs.NewEmptyConfig(), map[string]*terraform.InputValue{})

		runner.EmitIssue(tc.Rule, tc.Message, tc.Location)

		if !cmp.Equal(runner.Issues, tc.Expected) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(runner.Issues, tc.Expected))
		}
	}
}
