package tflint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_EvaluateExpr_string(t *testing.T) {
	cases := []struct {
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  string
	}{
		{
			Name:      "literal",
			Key:       "literal",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "literal_val",
		},
		{
			Name: "string interpolation",
			Key:  "string",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"string_var": cty.StringVal("string_val"),
				},
			},
			Expected: "string_val",
		},
		{
			Name: "new style interpolation",
			Key:  "new_string",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"string_var": cty.StringVal("string_val"),
				},
			},
			Expected: "string_val",
		},
		{
			Name: "list element",
			Key:  "list_element",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"list_var": cty.TupleVal([]cty.Value{cty.StringVal("one"), cty.StringVal("two")}),
				},
			},
			Expected: "one",
		},
		{
			Name: "map element",
			Key:  "map_element",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"map_var": cty.ObjectVal(map[string]cty.Value{"one": cty.StringVal("one"), "two": cty.StringVal("two")}),
				},
			},
			Expected: "one",
		},
		{
			Name: "convert from integer",
			Key:  "string",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"string_var": cty.NumberIntVal(10),
				},
			},
			Expected: "10",
		},
		{
			Name:      "conditional",
			Key:       "conditional",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "production",
		},
		{
			Name:      "bulit-in function",
			Key:       "function",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "acbd18db4cc2f85cedef654fccc4a4d8",
		},
		{
			Name:      "terraform workspace",
			Key:       "workspace",
			Variables: map[string]map[string]cty.Value{},
			Expected:  "default",
		},
		{
			Name: "inside interpolation",
			Key:  "inside",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"string_var": cty.StringVal("World"),
				},
			},
			Expected: "Hello World",
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Meta: &terraform.ContextMeta{
						Env: getWorkspace(),
					},
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

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
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  int
	}{
		{
			Name: "integer interpolation",
			Key:  "integer",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"integer_var": cty.NumberIntVal(3),
				},
			},
			Expected: 3,
		},
		{
			Name: "convert from string",
			Key:  "integer",
			Variables: map[string]map[string]cty.Value{
				"": map[string]cty.Value{
					"integer_var": cty.StringVal("3"),
				},
			},
			Expected: 3,
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Meta: &terraform.ContextMeta{
						Env: getWorkspace(),
					},
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

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
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  []string
	}{
		{
			Name:      "list literal",
			Key:       "string_list",
			Variables: map[string]map[string]cty.Value{},
			Expected:  []string{"one", "two", "three"},
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Meta: &terraform.ContextMeta{
						Env: getWorkspace(),
					},
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

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
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  []int
	}{
		{
			Name:      "list literal",
			Key:       "number_list",
			Variables: map[string]map[string]cty.Value{},
			Expected:  []int{1, 2, 3},
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Meta: &terraform.ContextMeta{
						Env: getWorkspace(),
					},
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

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
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		Expected  mapObject
	}{
		{
			Name:      "map literal",
			Key:       "map",
			Variables: map[string]map[string]cty.Value{},
			Expected:  mapObject{One: 1, Two: 2},
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Meta: &terraform.ContextMeta{
						Env: getWorkspace(),
					},
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

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
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name      string
		Key       string
		Variables map[string]map[string]cty.Value
		ErrorCode int
		ErrorText string
	}{
		{
			Name:      "undefined variable",
			Key:       "undefined",
			Variables: map[string]map[string]cty.Value{},
			ErrorCode: EvaluationError,
			ErrorText: fmt.Sprintf(
				"Failed to eval an expression in %s:33: Reference to undeclared input variable: An input variable with the name \"undefined_var\" has not been declared. This variable can be declared with a variable \"undefined_var\" {} block.",
				filepath.Join(currentDir, "test-fixtures", "runner", "resource.tf"),
			),
		},
		{
			Name:      "no default value",
			Key:       "no_value",
			Variables: map[string]map[string]cty.Value{},
			ErrorCode: UnknownValueError,
			ErrorText: fmt.Sprintf(
				"Unknown value found in %s:34; Please use environment variables or tfvars to set the value",
				filepath.Join(currentDir, "test-fixtures", "runner", "resource.tf"),
			),
		},
		{
			Name:      "terraform env",
			Key:       "env",
			Variables: map[string]map[string]cty.Value{},
			ErrorCode: EvaluationError,
			ErrorText: fmt.Sprintf(
				"Failed to eval an expression in %s:35: Invalid \"terraform\" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The \"state environment\" concept was rename to \"workspace\" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.",
				filepath.Join(currentDir, "test-fixtures", "runner", "resource.tf"),
			),
		},
		{
			Name:      "type mismatch",
			Key:       "string_list",
			Variables: map[string]map[string]cty.Value{},
			ErrorCode: TypeConversionError,
			ErrorText: fmt.Sprintf(
				"Invalid type expression in %s:23: incorrect type; string required",
				filepath.Join(currentDir, "test-fixtures", "runner", "resource.tf"),
			),
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		runner := &Runner{
			ctx: terraform.BuiltinEvalContext{
				Evaluator: &terraform.Evaluator{
					Meta: &terraform.ContextMeta{
						Env: getWorkspace(),
					},
					Config:             cfg,
					VariableValues:     tc.Variables,
					VariableValuesLock: &sync.Mutex{},
				},
			},
		}

		var ret string
		err = runner.EvaluateExpr(attribute.Expr, &ret)
		if appErr, ok := err.(*Error); ok {
			if appErr == nil {
				t.Fatalf("Failed `%s` test: expected err is `%s`, but nothing occurred", tc.Name, tc.ErrorText)
			}
			if appErr.Code != tc.ErrorCode {
				t.Fatalf("Failed `%s` test: expected error code is `%d`, but get `%d`", tc.Name, tc.ErrorCode, appErr.Code)
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
		Key      string
		Expected bool
	}{
		{
			Name:     "literal",
			Key:      "literal",
			Expected: true,
		},
		{
			Name:     "var syntax",
			Key:      "string",
			Expected: true,
		},
		{
			Name:     "new var syntax",
			Key:      "new_string",
			Expected: true,
		},
		{
			Name:     "conditional",
			Key:      "conditional",
			Expected: true,
		},
		{
			Name:     "function",
			Key:      "function",
			Expected: true,
		},
		{
			Name:     "terraform attributes",
			Key:      "workspace",
			Expected: true,
		},
		{
			Name:     "include supported syntax",
			Key:      "inside",
			Expected: true,
		},
		{
			Name:     "list",
			Key:      "string_list",
			Expected: true,
		},
		{
			Name:     "map",
			Key:      "map",
			Expected: true,
		},
		{
			Name:     "module",
			Key:      "module",
			Expected: false,
		},
		{
			Name:     "resource",
			Key:      "resource",
			Expected: false,
		},
		{
			Name:     "include unsupported syntax",
			Key:      "unsupported",
			Expected: false,
		},
	}

	for _, tc := range cases {
		cfg, err := loadConfigHelper()
		if err != nil {
			t.Fatal(err)
		}
		attribute, err := extractAttributeHelper(tc.Key, cfg)
		if err != nil {
			t.Fatal(err)
		}

		ret := isEvaluable(attribute.Expr)
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %t, but get %t", tc.Name, tc.Expected, ret)
		}
	}
}

func Test_getWorkspace(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Dir      string
		EnvVar   string
		Expected string
	}{
		{
			Name:     "default",
			Expected: "default",
		},
		{
			Name:     "environment variable",
			EnvVar:   "dev",
			Expected: "dev",
		},
		{
			Name:     "environment file",
			Dir:      filepath.Join(currentDir, "test-fixtures", "runner", "environment"),
			Expected: "staging",
		},
	}

	for _, tc := range cases {
		if tc.Dir != "" {
			err := os.Chdir(tc.Dir)
			if err != nil {
				t.Fatal(err)
			}
		}

		if tc.EnvVar != "" {
			err := os.Setenv("TF_WORKSPACE", tc.EnvVar)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getWorkspace()
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		if tc.EnvVar != "" {
			err := os.Unsetenv("TF_WORKSPACE")
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

func loadConfigHelper() (*configs.Config, error) {
	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		return nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	mod, diags := loader.Parser().LoadConfigDir(dir + "/test-fixtures/runner")
	if diags.HasErrors() {
		return nil, diags
	}
	cfg, diags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
	if diags.HasErrors() {
		return nil, diags
	}

	return cfg, nil
}

func extractAttributeHelper(key string, cfg *configs.Config) (*hcl.Attribute, error) {
	resource := cfg.Module.ManagedResources["null_resource.test"]
	if resource == nil {
		return nil, errors.New("Expected resource is not found")
	}
	body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: key,
			},
		},
	})
	if diags.HasErrors() {
		return nil, diags
	}
	attribute := body.Attributes[key]
	if attribute == nil {
		return nil, fmt.Errorf("Expected attribute is not found: %s", key)
	}
	return attribute, nil
}
