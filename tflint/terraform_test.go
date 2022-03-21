package tflint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_ParseTFVariables(t *testing.T) {
	cases := []struct {
		Name     string
		DeclVars map[string]*configs.Variable
		Vars     []string
		Expected terraform.InputValues
	}{
		{
			Name:     "undeclared",
			DeclVars: map[string]*configs.Variable{},
			Vars: []string{
				"foo=bar",
				"bar=[\"foo\"]",
				"baz={ foo=\"bar\" }",
			},
			Expected: terraform.InputValues{
				"foo": &terraform.InputValue{
					Value:      cty.StringVal("bar"),
					SourceType: terraform.ValueFromCLIArg,
				},
				"bar": &terraform.InputValue{
					Value:      cty.StringVal("[\"foo\"]"),
					SourceType: terraform.ValueFromCLIArg,
				},
				"baz": &terraform.InputValue{
					Value:      cty.StringVal("{ foo=\"bar\" }"),
					SourceType: terraform.ValueFromCLIArg,
				},
			},
		},
		{
			Name: "declared",
			DeclVars: map[string]*configs.Variable{
				"foo": {ParsingMode: configs.VariableParseLiteral},
				"bar": {ParsingMode: configs.VariableParseHCL},
				"baz": {ParsingMode: configs.VariableParseHCL},
			},
			Vars: []string{
				"foo=bar",
				"bar=[\"foo\"]",
				"baz={ foo=\"bar\" }",
			},
			Expected: terraform.InputValues{
				"foo": &terraform.InputValue{
					Value:      cty.StringVal("bar"),
					SourceType: terraform.ValueFromCLIArg,
				},
				"bar": &terraform.InputValue{
					Value:      cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
					SourceType: terraform.ValueFromCLIArg,
				},
				"baz": &terraform.InputValue{
					Value:      cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
					SourceType: terraform.ValueFromCLIArg,
				},
			},
		},
	}

	for _, tc := range cases {
		ret, err := ParseTFVariables(tc.Vars, tc.DeclVars)
		if err != nil {
			t.Fatalf("Failed `%s` test: Unexpected error occurred: %s", tc.Name, err)
		}

		if !reflect.DeepEqual(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test:\n Expected: %#v\n Actual: %#v", tc.Name, tc.Expected, ret)
		}
	}
}

func Test_ParseTFVariables_errors(t *testing.T) {
	cases := []struct {
		Name     string
		DeclVars map[string]*configs.Variable
		Vars     []string
		Expected string
	}{
		{
			Name:     "invalid format",
			DeclVars: map[string]*configs.Variable{},
			Vars:     []string{"foo"},
			Expected: "`foo` is invalid. Variables must be `key=value` format",
		},
		{
			Name: "invalid parsing mode",
			DeclVars: map[string]*configs.Variable{
				"foo": {ParsingMode: configs.VariableParseHCL},
			},
			Vars:     []string{"foo=bar"},
			Expected: "<value for var.foo>:1,1-4: Variables not allowed; Variables may not be used here.",
		},
		{
			Name: "invalid expression",
			DeclVars: map[string]*configs.Variable{
				"foo": {ParsingMode: configs.VariableParseHCL},
			},
			Vars:     []string{"foo="},
			Expected: "<value for var.foo>:1,1-1: Missing expression; Expected the start of an expression, but found the end of the file.",
		},
	}

	for _, tc := range cases {
		_, err := ParseTFVariables(tc.Vars, tc.DeclVars)
		if err == nil {
			t.Fatalf("Failed `%s` test: Expected an error, but nothing occurred", tc.Name)
		}

		if err.Error() != tc.Expected {
			t.Fatalf("Failed `%s` test: Expected `%s`, but got `%s`", tc.Name, tc.Expected, err.Error())
		}
	}
}

func Test_getTFDataDir(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: ".terraform",
		},
		{
			Name:     "environment variable",
			EnvVar:   map[string]string{"TF_DATA_DIR": ".tfdata"},
			Expected: ".tfdata",
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFDataDir()
		if ret != tc.Expected {
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

func Test_getTFModuleDir(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: filepath.Join(".terraform", "modules"),
		},
		{
			Name:     "environment variable",
			EnvVar:   map[string]string{"TF_DATA_DIR": ".tfdata"},
			Expected: filepath.Join(".tfdata", "modules"),
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFModuleDir()
		if ret != tc.Expected {
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

func Test_getTFModuleManifestPath(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: filepath.Join(".terraform", "modules", "modules.json"),
		},
		{
			Name:     "environment variable",
			EnvVar:   map[string]string{"TF_DATA_DIR": ".tfdata"},
			Expected: filepath.Join(".tfdata", "modules", "modules.json"),
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFModuleManifestPath()
		if ret != tc.Expected {
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

func Test_getTFWorkspace(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Dir      string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: "default",
		},
		{
			Name:     "TF_WORKSPACE",
			EnvVar:   map[string]string{"TF_WORKSPACE": "dev"},
			Expected: "dev",
		},
		{
			Name:     "environment file",
			Dir:      filepath.Join(currentDir, "test-fixtures", "with_environment_file"),
			Expected: "staging",
		},
		{
			Name:     "TF_DATA_DIR",
			Dir:      filepath.Join(currentDir, "test-fixtures", "with_environment_file"),
			EnvVar:   map[string]string{"TF_DATA_DIR": ".terraform_production"},
			Expected: "production",
		},
	}

	for _, tc := range cases {
		if tc.Dir != "" {
			err := os.Chdir(tc.Dir)
			if err != nil {
				t.Fatal(err)
			}
		}

		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFWorkspace()
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}

		if tc.Dir != "" {
			err := os.Chdir(currentDir)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func Test_getTFEnvVariables(t *testing.T) {
	cases := []struct {
		Name     string
		DeclVars map[string]*configs.Variable
		EnvVar   map[string]string
		Expected terraform.InputValues
	}{
		{
			Name:     "undeclared",
			DeclVars: map[string]*configs.Variable{},
			EnvVar: map[string]string{
				"TF_VAR_instance_type": "t2.micro",
				"TF_VAR_count":         "5",
				"TF_VAR_list":          "[\"foo\"]",
				"TF_VAR_map":           "{foo=\"bar\"}",
			},
			Expected: terraform.InputValues{
				"instance_type": &terraform.InputValue{
					Value:      cty.StringVal("t2.micro"),
					SourceType: terraform.ValueFromEnvVar,
				},
				"count": &terraform.InputValue{
					Value:      cty.StringVal("5"),
					SourceType: terraform.ValueFromEnvVar,
				},
				"list": &terraform.InputValue{
					Value:      cty.StringVal("[\"foo\"]"),
					SourceType: terraform.ValueFromEnvVar,
				},
				"map": &terraform.InputValue{
					Value:      cty.StringVal("{foo=\"bar\"}"),
					SourceType: terraform.ValueFromEnvVar,
				},
			},
		},
		{
			Name: "declared",
			DeclVars: map[string]*configs.Variable{
				"instance_type": {ParsingMode: configs.VariableParseLiteral},
				"count":         {ParsingMode: configs.VariableParseHCL},
				"list":          {ParsingMode: configs.VariableParseHCL},
				"map":           {ParsingMode: configs.VariableParseHCL},
			},
			EnvVar: map[string]string{
				"TF_VAR_instance_type": "t2.micro",
				"TF_VAR_count":         "5",
				"TF_VAR_list":          "[\"foo\"]",
				"TF_VAR_map":           "{foo=\"bar\"}",
			},
			Expected: terraform.InputValues{
				"instance_type": &terraform.InputValue{
					Value:      cty.StringVal("t2.micro"),
					SourceType: terraform.ValueFromEnvVar,
				},
				"count": &terraform.InputValue{
					Value:      cty.NumberIntVal(5),
					SourceType: terraform.ValueFromEnvVar,
				},
				"list": &terraform.InputValue{
					Value:      cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
					SourceType: terraform.ValueFromEnvVar,
				},
				"map": &terraform.InputValue{
					Value:      cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
					SourceType: terraform.ValueFromEnvVar,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for key, value := range tc.EnvVar {
				t.Setenv(key, value)
			}

			ret, diags := getTFEnvVariables(tc.DeclVars)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opt := cmp.Comparer(func(x, y cty.Value) bool {
				return x.RawEquals(y)
			})
			if diff := cmp.Diff(tc.Expected, ret, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_getTFEnvVariables_errors(t *testing.T) {
	cases := []struct {
		Name     string
		DeclVars map[string]*configs.Variable
		Env      map[string]string
		Expected string
	}{
		{
			Name: "invalid parsing mode",
			DeclVars: map[string]*configs.Variable{
				"foo": {ParsingMode: configs.VariableParseHCL},
			},
			Env: map[string]string{
				"TF_VAR_foo": "bar",
			},
			Expected: "<value for var.foo>:1,1-4: Variables not allowed; Variables may not be used here.",
		},
		{
			Name: "invalid expression",
			DeclVars: map[string]*configs.Variable{
				"foo": {ParsingMode: configs.VariableParseHCL},
			},
			Env: map[string]string{
				"TF_VAR_foo": `{"bar": "baz"`,
			},
			Expected: "<value for var.foo>:1,1-2: Unterminated object constructor expression; There is no corresponding closing brace before the end of the file. This may be caused by incorrect brace nesting elsewhere in this file.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for k, v := range tc.Env {
				t.Setenv(k, v)
			}

			_, diags := getTFEnvVariables(tc.DeclVars)
			if !diags.HasErrors() {
				t.Fatal("Expected an error to occur, but it didn't")
			}

			if diags.Error() != tc.Expected {
				t.Errorf("Expected `%s`, but got `%s`", tc.Expected, diags.Error())
			}
		})
	}
}
