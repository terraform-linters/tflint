package terraform

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

func TestPartialContent(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string]string
		schema *hclext.BodySchema
		want   *hclext.BodyContent
	}{
		{
			name:  "empty files",
			files: map[string]string{},
			schema: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
					},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			name: "primaries",
			files: map[string]string{
				"main1.tf": `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
				"main2.tf": `
resource "aws_instance" "bar" {
  instance_type = "m5.2xlarge"
}`,
			},
			schema: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main1.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main1.tf"},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main2.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main2.tf"},
					},
				},
			},
		},
		{
			name: "overrides",
			files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
				"main_override.tf": `
resource "aws_instance" "foo" {
  instance_type = "m5.2xlarge"
}`,
			},
			schema: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main_override.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			name: "just attributes",
			files: map[string]string{
				"main.tf": `
locals {
  foo = "foo"
  bar = "bar"
}`,
			},
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "locals",
						Body: &hclext.BodySchema{Mode: hclext.SchemaJustAttributesMode},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type: "locals",
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{
								"foo": &hclext.Attribute{Name: "foo", Range: hcl.Range{Filename: "main.tf"}},
								"bar": &hclext.Attribute{Name: "bar", Range: hcl.Range{Filename: "main.tf"}},
							},
							Blocks: hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			name: "expand resources",
			files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
  count = 0
  instance_type = "t2.micro"
}
resource "aws_instance" "bar" {
  count = 2
  instance_type = "m5.2xlarge"
}`,
			},
			schema: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			name: "expand modules",
			files: map[string]string{
				"main.tf": `
module "foo" {
  count = 0
  instance_type = "t2.micro"
}
module "bar" {
  count = 2
  instance_type = "m5.2xlarge"
}`,
			},
			schema: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type:       "module",
						LabelNames: []string{"name"},
						Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "instance_type"}}},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "module",
						Labels: []string{"bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
					{
						Type:   "module",
						Labels: []string{"bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf"}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			name: "dynamic blocks",
			files: map[string]string{
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
			schema: &hclext.BodySchema{
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
			},
			want: &hclext.BodyContent{
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
										Blocks:     hclext.Blocks{},
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
										Blocks:     hclext.Blocks{},
									},
									DefRange: hcl.Range{Filename: "main.tf"},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"volume_size": &hclext.Attribute{Name: "volume_size", Range: hcl.Range{Filename: "main.tf"}}},
										Blocks:     hclext.Blocks{},
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
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, content := range test.files {
				if err := fs.WriteFile(name, []byte(content), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}

			parser := NewParser(fs)
			mod, diags := parser.LoadConfigDir(".", ".")
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			config, diags := BuildConfig(mod, ModuleWalkerFunc(func(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) { return nil, nil, nil }))
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			variableValues, diags := VariableValues(config)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := &Evaluator{
				Meta:           &ContextMeta{Env: Workspace()},
				ModulePath:     config.Path.UnkeyedInstanceShim(),
				Config:         config,
				VariableValues: variableValues,
			}

			got, diags := config.Module.PartialContent(test.schema, ctx)
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
			if diff := cmp.Diff(got, test.want, opts); diff != "" {
				t.Error(diff)
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
