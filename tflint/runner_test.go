package tflint

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
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

		expectedVars := map[string]map[string]*configs.Variable{
			"module.root": {
				"override": {
					Name:        "override",
					Default:     cty.StringVal("foo"),
					Type:        cty.DynamicPseudoType,
					ParsingMode: configs.VariableParseLiteral,
					Nullable:    true,
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
					Nullable:    true,
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
					Nullable:    true,
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
					ParsingMode: configs.VariableParseLiteral,
					Nullable:    true,
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
					Nullable:    true,
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
					Nullable:    true,
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

		expected := errors.New("failed to eval an expression in module.tf:4; Invalid \"terraform\" attribute: The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The \"state environment\" concept was renamed to \"workspace\" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.")
		if err == nil {
			t.Fatal("an error was expected to occur, but it did not")
		}
		if expected.Error() != err.Error() {
			t.Fatalf("expected error is `%s`, but get `%s`", expected, err)
		}
	})
}

func Test_NewModuleRunners_withNotAllowedAttributes(t *testing.T) {
	withinFixtureDir(t, "not_allowed_module_attribute", func() {
		runner := testRunnerWithOsFs(t, moduleConfig())

		_, err := NewModuleRunners(runner)

		expected := errors.New("attribute of module not allowed was found in module.tf:1; module.tf:4,3-10: Unexpected \"invalid\" block; Blocks are not allowed here.")
		if err == nil {
			t.Fatal("an error was expected to occur, but it did not")
		}
		if expected.Error() != err.Error() {
			t.Fatalf("expected error is `%s`, but get `%s`", expected, err)
		}
	})
}

func TestGetModuleContent(t *testing.T) {
	tests := []struct {
		Name  string
		Files map[string]string
		Args  func() (*hclext.BodySchema, sdk.GetModuleContentOption)
		Want  *hclext.BodyContent
	}{
		{
			Name:  "empty files",
			Files: map[string]string{},
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
						},
					},
				}, sdk.GetModuleContentOption{Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{},
		},
		{
			Name: "primaries",
			Files: map[string]string{
				"main1.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`,
				"main2.tf": `
resource "aws_instance" "bar" {
	instance_type = "m5.2xlarge"
}`,
			},
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
						},
					},
				}, sdk.GetModuleContentOption{Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main1.tf"}}},
						},
						DefRange: hcl.Range{Filename: "main1.tf"},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main2.tf"}}},
						},
						DefRange: hcl.Range{Filename: "main2.tf"},
					},
				},
			},
		},
		{
			Name: "overrides",
			Files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
	instance_type = "t2.micro"
}`,
				"main_override.tf": `
resource "aws_instance" "foo" {
	instance_type = "m5.2xlarge"
}`,
			},
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
						},
					},
				}, sdk.GetModuleContentOption{Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main_override.tf"}}},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			Name: "contains not created resource",
			Files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
	count = 0
	instance_type = "t2.micro"
}

resource "aws_instance" "bar" {
	count = 1
	instance_type = "m5.2xlarge"
}`,
			},
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
						},
					},
				}, sdk.GetModuleContentOption{Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf"}}},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			Name: "dynamic blocks",
			Files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
	ebs_block_device {
		volume_size = "10"
	}
}

resource "aws_instance" "bar" {
	dynamic "ebs_block_device" {
		for_each = toset([20, 30])
		content {
			volume_size = ebs_block_device.value
		}
	}
}`,
			},
			Args: func() (*hclext.BodySchema, sdk.GetModuleContentOption) {
				return &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "foo"}},
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
								Blocks: []hclext.BlockSchema{
									{
										Type: "ebs_block_device",
										Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "volume_size"}}},
									},
								},
							},
						},
					},
				}, sdk.GetModuleContentOption{Hint: sdk.GetModuleContentHint{ResourceType: "aws_instance"}}
			},
			Want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"volume_size": &hclext.Attribute{Name: "volume_size", Range: hcl.Range{Filename: "main.tf"}}},
									},
									DefRange: hcl.Range{Filename: "main.tf"},
								},
							},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"volume_size": &hclext.Attribute{Name: "volume_size", Range: hcl.Range{Filename: "main.tf"}}},
									},
									DefRange: hcl.Range{Filename: "main.tf"},
								},
							},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := TestRunner(t, test.Files)

			got, diags := runner.GetModuleContent(test.Args())
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hclext.Block{}, "TypeRange", "LabelRanges"),
				cmpopts.IgnoreFields(hclext.Attribute{}, "Expr", "NameRange"),
				cmpopts.IgnoreFields(hcl.Range{}, "Start", "End"),
				cmpopts.SortSlices(func(i, j *hclext.Block) bool {
					return i.DefRange.String() < j.DefRange.String()
				}),
			}
			if diff := cmp.Diff(got, test.Want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func Test_appendDynamicBlockSchema(t *testing.T) {
	tests := []struct {
		Name string
		In   *hclext.BodySchema
		Want *hclext.BodySchema
	}{
		{
			Name: "empty schema",
			In:   &hclext.BodySchema{},
			Want: &hclext.BodySchema{},
		},
		{
			Name: "attribute schemas",
			In: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
			Want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
		},
		{
			Name: "block schemas",
			In: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type: "toplevel",
						Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
					},
				},
			},
			Want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type: "toplevel",
						Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
					},
					{
						Type:       "dynamic",
						LabelNames: []string{"name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "content",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "nested block schemas",
			In: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type: "toplevel",
						Body: &hclext.BodySchema{
							Attributes: []hclext.AttributeSchema{{Name: "bar"}},
							Blocks: []hclext.BlockSchema{
								{
									Type: "nested",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
								},
							},
						},
					},
				},
			},
			Want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type: "toplevel",
						Body: &hclext.BodySchema{
							Attributes: []hclext.AttributeSchema{{Name: "bar"}},
							Blocks: []hclext.BlockSchema{
								{
									Type: "nested",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
								},
								{
									Type:       "dynamic",
									LabelNames: []string{"name"},
									Body: &hclext.BodySchema{
										Blocks: []hclext.BlockSchema{
											{
												Type: "content",
												Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
											},
										},
									},
								},
							},
						},
					},
					{
						Type:       "dynamic",
						LabelNames: []string{"name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "content",
									Body: &hclext.BodySchema{
										Attributes: []hclext.AttributeSchema{{Name: "bar"}},
										Blocks: []hclext.BlockSchema{
											{
												Type: "nested",
												Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
											},
											{
												Type:       "dynamic",
												LabelNames: []string{"name"},
												Body: &hclext.BodySchema{
													Blocks: []hclext.BlockSchema{
														{
															Type: "content",
															Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := appendDynamicBlockSchema(test.In)

			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func Test_resolveDynamicBlocks(t *testing.T) {
	tests := []struct {
		Name string
		In   *hclext.BodyContent
		Want *hclext.BodyContent
	}{
		{
			Name: "empty body",
			In:   &hclext.BodyContent{},
			Want: &hclext.BodyContent{},
		},
		{
			Name: "only attributes",
			In: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
			},
		},
		{
			Name: "regular blocks",
			In: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type: "toplevel",
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
						},
					},
				},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type: "toplevel",
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
						},
					},
				},
			},
		},
		{
			Name: "dynamic blocks",
			In: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type:   "dynamic",
						Labels: []string{"toplevel"},
						Body: &hclext.BodyContent{
							Blocks: hclext.Blocks{
								{
									Type: "content",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
									},
								},
							},
						},
					},
				},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type: "toplevel",
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
						},
					},
				},
			},
		},
		{
			Name: "dynamic nested blocks in regular blocks",
			In: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type: "toplevel",
						Body: &hclext.BodyContent{
							Blocks: hclext.Blocks{
								{
									Type:   "dynamic",
									Labels: []string{"nested"},
									Body: &hclext.BodyContent{
										Blocks: hclext.Blocks{
											{
												Type: "content",
												Body: &hclext.BodyContent{
													Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type: "toplevel",
						Body: &hclext.BodyContent{
							Blocks: hclext.Blocks{
								{
									Type: "nested",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "dynamic nested blocks in dynamic blocks",
			In: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type:   "dynamic",
						Labels: []string{"toplevel"},
						Body: &hclext.BodyContent{
							Blocks: hclext.Blocks{
								{
									Type: "content",
									Body: &hclext.BodyContent{
										Blocks: hclext.Blocks{
											{
												Type:   "dynamic",
												Labels: []string{"nested"},
												Body: &hclext.BodyContent{
													Blocks: hclext.Blocks{
														{
															Type: "content",
															Body: &hclext.BodyContent{
																Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
				Blocks: hclext.Blocks{
					{
						Type: "toplevel",
						Body: &hclext.BodyContent{
							Blocks: hclext.Blocks{
								{
									Type: "nested",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := resolveDynamicBlocks(test.In)

			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func Test_overrideBlocks(t *testing.T) {
	tests := []struct {
		Name      string
		Primaries hclext.Blocks
		Overrides hclext.Blocks
		Want      hclext.Blocks
	}{
		{
			Name:      "empty blocks",
			Primaries: hclext.Blocks{},
			Overrides: hclext.Blocks{},
			Want:      hclext.Blocks{},
		},
		{
			Name: "no override",
			Primaries: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
					},
				},
			},
			Overrides: hclext.Blocks{},
			Want: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
					},
				},
			},
		},
		{
			Name: "override",
			Primaries: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "foo"},
							"bar": &hclext.Attribute{Name: "bar"},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "bar"},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "bar"},
							"bar": &hclext.Attribute{Name: "bar"},
						},
					},
				},
			},
		},
		{
			Name: "override nested blocks",
			Primaries: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "baz"},
										"qux": &hclext.Attribute{Name: "qux"},
									},
								},
							},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "bar"}},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "qux"},
									},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "resource",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "bar"}},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "qux"},
										"qux": &hclext.Attribute{Name: "qux"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := overrideBlocks(test.Primaries, test.Overrides)

			if diff := cmp.Diff(got, test.Want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func Test_RunnerFiles(t *testing.T) {
	runner := TestRunner(t, map[string]string{
		"main.tf": "",
	})
	runner.files["child/main.tf"] = &hcl.File{}

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

func Test_DecodeRuleConfig(t *testing.T) {
	type ruleSchema struct {
		Foo string `hcl:"foo"`
	}
	options := ruleSchema{}

	file, diags := hclsyntax.ParseConfig([]byte(`foo = "bar"`), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		t.Fatalf("Failed to parse test config: %s", diags)
	}

	cfg := EmptyConfig()
	cfg.Rules["test"] = &RuleConfig{
		Name:    "test",
		Enabled: true,
		Body:    file.Body,
	}

	runner := TestRunnerWithConfig(t, map[string]string{}, cfg)
	if err := runner.DecodeRuleConfig("test", &options); err != nil {
		t.Fatalf("Failed to decode rule config: %s", err)
	}

	expected := ruleSchema{Foo: "bar"}
	if !cmp.Equal(options, expected) {
		t.Fatalf("Failed to decode rule config: diff=%s", cmp.Diff(options, expected))
	}
}

func Test_DecodeRuleConfig_emptyBody(t *testing.T) {
	type ruleSchema struct {
		Foo string `hcl:"foo"`
	}
	options := ruleSchema{}

	cfg := EmptyConfig()
	cfg.Rules["test"] = &RuleConfig{
		Name:    "test",
		Enabled: true,
		Body:    hcl.EmptyBody(),
	}

	runner := TestRunnerWithConfig(t, map[string]string{}, cfg)
	err := runner.DecodeRuleConfig("test", &options)
	if err == nil {
		t.Fatal("Expected to fail to decode rule config, but not")
	}

	expected := "This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration"
	if err.Error() != expected {
		t.Fatalf("Expected error message is %s, but got %s", expected, err.Error())
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
