package tflint

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/zclconf/go-cty/cty"
)

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

		expectedVars := map[string]map[string]*terraform.Variable{
			"module.root": {
				"override": {
					Name:        "override",
					Default:     cty.StringVal("foo"),
					Type:        cty.DynamicPseudoType,
					Nullable:    true,
					ParsingMode: terraform.VariableParseLiteral,
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
					Nullable:    true,
					ParsingMode: terraform.VariableParseLiteral,
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
					Nullable:    true,
					ParsingMode: terraform.VariableParseLiteral,
					DeclRange: hcl.Range{
						Filename: filepath.Join("module", "module.tf"),
						Start:    hcl.Pos{Line: 5, Column: 1},
						End:      hcl.Pos{Line: 5, Column: 19},
					},
				},
			},
			"module.root.module.test": {
				"override": {
					Name:        "override",
					Default:     cty.StringVal("foo"),
					Type:        cty.DynamicPseudoType,
					Nullable:    true,
					ParsingMode: terraform.VariableParseLiteral,
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
					Nullable:    true,
					ParsingMode: terraform.VariableParseLiteral,
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
					Nullable:    true,
					ParsingMode: terraform.VariableParseLiteral,
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

func Test_NewModuleRunners_withCountForEach(t *testing.T) {
	withinFixtureDir(t, "module_with_count_for_each", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		runners, err := NewModuleRunners(runner)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if len(runners) != 5 {
			t.Fatalf("This function must return 5 runners, but returned %d", len(runners))
		}

		moduleNames := make([]string, 5)
		for idx, r := range runners {
			moduleNames[idx] = r.TFConfig.Path.String()
		}
		expected := []string{
			"module.count_is_one",
			"module.count_is_two",
			"module.count_is_two",
			"module.for_each_is_not_empty",
			"module.for_each_is_not_empty",
		}
		less := func(a, b string) bool { return a < b }
		if diff := cmp.Diff(moduleNames, expected, cmpopts.SortSlices(less)); diff != "" {
			t.Fatal(diff)
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
		if child.TFConfig.Path.String() != "module.module1" {
			t.Fatalf("Expected child config path name is `module.module1`, but get `%s`", child.TFConfig.Path.String())
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
		if grandchild.TFConfig.Path.String() != "module.module1.module.module2" {
			t.Fatalf("Expected child config path name is `module.module1.module.module2`, but get `%s`", grandchild.TFConfig.Path.String())
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
		opts = []cmp.Option{
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
			cmpopts.SortSlices(func(x, y *moduleVariable) bool { return x.DeclRange.Start.Line > y.DeclRange.Start.Line }),
		}
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

		expected := errors.New(`module.tf:4,16-29: Invalid "terraform" attribute; The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The "state environment" concept was renamed to "workspace" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.`)
		if err == nil {
			t.Fatal("an error was expected to occur, but it did not")
		}
		if expected.Error() != err.Error() {
			t.Fatalf("expected error is `%s`, but get `%s`", expected, err)
		}
	})
}

func Test_RunnerFiles(t *testing.T) {
	runner := TestRunner(t, map[string]string{
		"main.tf": "",
	})
	runner.TFConfig.Module.Files["child/main.tf"] = &hcl.File{}

	expected := map[string]*hcl.File{
		"main.tf": {
			Body:  hcl.EmptyBody(),
			Bytes: []byte{},
		},
	}

	files := runner.Files()

	opt := cmpopts.IgnoreFields(hcl.File{}, "Body", "Nav")
	if !cmp.Equal(expected, files, opt) {
		t.Fatalf("Failed test: diff: %s", cmp.Diff(expected, files, opt))
	}
}

func Test_LookupIssues(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		issues   Issues
		expected Issues
	}{
		{
			name: "multiple files",
			in:   "template.tf",
			issues: Issues{
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
			},
			expected: Issues{
				{
					Rule:    &testRule{},
					Message: "This is test rule",
					Range: hcl.Range{
						Filename: "template.tf",
						Start:    hcl.Pos{Line: 1},
					},
				},
			},
		},
		{
			name: "path normalization",
			in:   "./template.tf",
			issues: Issues{
				{
					Rule:    &testRule{},
					Message: "This is test rule",
					Range: hcl.Range{
						Filename: "template.tf",
						Start:    hcl.Pos{Line: 1},
					},
				},
			},
			expected: Issues{
				{
					Rule:    &testRule{},
					Message: "This is test rule",
					Range: hcl.Range{
						Filename: "template.tf",
						Start:    hcl.Pos{Line: 1},
					},
				},
			},
		},
	}

	runner := TestRunner(t, map[string]string{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner.Issues = test.issues

			got := runner.LookupIssues(test.in)
			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

type testRule struct{}

func (r *testRule) Name() string {
	return "test_rule"
}
func (r *testRule) Severity() Severity {
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

func Test_listVarRefs(t *testing.T) {
	cases := []struct {
		Name     string
		Expr     string
		Expected map[string]addrs.InputVariable
	}{
		{
			Name:     "literal",
			Expr:     "1",
			Expected: map[string]addrs.InputVariable{},
		},
		{
			Name: "input variable",
			Expr: "var.foo",
			Expected: map[string]addrs.InputVariable{
				"var.foo": {Name: "foo"},
			},
		},
		{
			Name:     "local variable",
			Expr:     "local.bar",
			Expected: map[string]addrs.InputVariable{},
		},
		{
			Name: "multiple input variables",
			Expr: `format("Hello, %s %s!", var.first_name, var.last_name)`,
			Expected: map[string]addrs.InputVariable{
				"var.first_name": {Name: "first_name"},
				"var.last_name":  {Name: "last_name"},
			},
		},
		{
			Name: "map input variable",
			Expr: `{
  name = var.tags["name"]
  env  = var.tags["env"]
}`,
			Expected: map[string]addrs.InputVariable{
				"var.tags": {Name: "tags"},
			},
		},
	}

	for _, tc := range cases {
		expr, diags := hclsyntax.ParseExpression([]byte(tc.Expr), "template.tf", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		refs := listVarRefs(expr)

		opt := cmpopts.IgnoreUnexported(addrs.InputVariable{})
		if !cmp.Equal(tc.Expected, refs, opt) {
			t.Fatalf("%s: Diff=%s", tc.Name, cmp.Diff(tc.Expected, refs, opt))
		}
	}
}
