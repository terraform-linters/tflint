package tflint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/terraform"
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
				"foo": &configs.Variable{ParsingMode: configs.VariableParseLiteral},
				"bar": &configs.Variable{ParsingMode: configs.VariableParseHCL},
				"baz": &configs.Variable{ParsingMode: configs.VariableParseHCL},
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
				"foo": &configs.Variable{ParsingMode: configs.VariableParseHCL},
			},
			Vars:     []string{"foo=bar"},
			Expected: "<value for var.foo>:1,1-4: Variables not allowed; Variables may not be used here.",
		},
		{
			Name: "invalid expression",
			DeclVars: map[string]*configs.Variable{
				"foo": &configs.Variable{ParsingMode: configs.VariableParseHCL},
			},
			Vars:     []string{"foo="},
			Expected: "<value for var.foo>:1,1-1: Invalid expression; Expected the start of an expression, but found an invalid expression token.",
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

func Test_HCLBodyRange_HCL(t *testing.T) {
	src := `
ebs_block_device {
  device_name = "/dev/sdf"
  volume_size = 10
  foo {
    bar = "baz"
  }
}
`

	file, diags := hclsyntax.ParseConfig([]byte(src), "example.tf", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	body, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "ebs_block_device",
			},
		},
	})
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	block := body.Blocks[0]

	got := HCLBodyRange(block.Body, block.DefRange)
	expected := hcl.Range{
		Filename: "example.tf",
		Start:    hcl.Pos{Line: 3, Column: 3},
		End:      hcl.Pos{Line: 7, Column: 4},
	}

	opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if !cmp.Equal(got, expected, opt) {
		t.Fatalf("Diff=%s", cmp.Diff(got, expected, opt))
	}
}

func Test_HCLBodyRange_JSON(t *testing.T) {
	src := `
{
  "ebs_block_device": {
    "device_name": "/dev/sdf",
    "volume_size": 10,
    "foo": {
      "bar": "baz"
    }
  }
}
`

	file, diags := json.Parse([]byte(src), "example.tf.json")
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	body, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "ebs_block_device",
			},
		},
	})
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	block := body.Blocks[0]

	got := HCLBodyRange(block.Body, block.DefRange)
	expected := hcl.Range{
		Filename: "example.tf.json",
		Start:    hcl.Pos{Line: 3, Column: 23},
		End:      hcl.Pos{Line: 9, Column: 4},
	}

	opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if !cmp.Equal(got, expected, opt) {
		t.Fatalf("Diff=%s", cmp.Diff(got, expected, opt))
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
		EnvVar   map[string]string
		Expected terraform.InputValues
	}{
		{
			Name: "environment variable",
			EnvVar: map[string]string{
				"TF_VAR_instance_type": "t2.micro",
				"TF_VAR_count":         "5",
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
			},
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFEnvVariables()
		if !reflect.DeepEqual(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test:\n Expected: %#v\n Actual: %#v", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
