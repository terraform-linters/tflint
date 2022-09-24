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
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			name: "contains not created resource",
			files: map[string]string{
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
						},
						DefRange: hcl.Range{Filename: "main.tf"},
					},
				},
			},
		},
		{
			name: "contains not created module",
			files: map[string]string{
				"main.tf": `
module "not_created" {
  count = 0
  instance_type = "t2.micro"
}
module "created" {
  count = 1
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
						Labels: []string{"created"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf"}}},
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
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, content := range test.files {
				if err := fs.WriteFile(name, []byte(content), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}

			parser := NewParser(fs)
			mod, diags := parser.LoadConfigDir(".")
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

func Test_expandBlocks(t *testing.T) {
	tests := []struct {
		name   string
		config string
		schema *hclext.BodySchema
		want   *hclext.BodyContent
	}{
		{
			name: "no meta-arguments",
			config: `
resource "aws_instance" "main" {}

module "aws_instance" {}
`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "count is not zero (literal)",
			config: `
resource "aws_instance" "main" {
  count = 1
}
module "aws_instance" {
  count = 1
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "count is not zero (variable)",
			config: `
variable "count" {
  default = 1
}
resource "aws_instance" "main" {
  count = var.count
}
module "aws_instance" {
  count = var.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "count is unknown",
			config: `
variable "count" {}

resource "aws_instance" "main" {
  count = var.count
}
module "aws_instance" {
  count = var.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			name: "count is sensitive",
			config: `
variable "count" {
  sensitive = true
  default   = 1
}
resource "aws_instance" "main" {
  count = var.count
}
module "aws_instance" {
  count = var.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "count is unevaluable",
			config: `
resource "aws_instance" "main" {
  count = module.meta.count
}
module "aws_instance" {
  count = module.meta.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			name: "count is zero",
			config: `
resource "aws_instance" "main" {
  count = 0
}
module "aws_instance" {
  count = 0
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			// HINT: Terraform does not allow null as `count`
			name: "count is null",
			config: `
resource "aws_instance" "main" {
  count = null
}
module "aws_instance" {
  count = null
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "for_each is not empty (literal)",
			config: `
resource "aws_instance" "main" {
  for_each = { foo = "bar" }
}
module "aws_instance" {
  for_each = { foo = "bar" }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "for_each is not empty (variable)",
			config: `
variable "for_each" {
  default = { foo = "bar" }
}
resource "aws_instance" "main" {
  for_each = var.for_each
}
module "aws_instance" {
  for_each = var.for_each
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "for_each is unknown",
			config: `
variable "for_each" {}

resource "aws_instance" "main" {
  for_each = var.for_each
}
module "aws_instance" {
  for_each = var.for_each
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			name: "for_each is evaluable",
			config: `
resource "aws_instance" "main" {
  for_each = module.meta.for_each
}
module "aws_instance" {
  for_each = module.meta.for_each
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			name: "for_each contains unevaluable",
			config: `
resource "aws_instance" "main" {
  for_each = {
    known   = "known"
    unknown = module.meta.unknown
  }
}
module "aws_instance" {
  for_each = {
    known   = "known"
    unknown = module.meta.unknown
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "for_each is empty",
			config: `
resource "aws_instance" "main" {
  for_each = {}
}
module "aws_instance" {
  for_each = {}
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			name: "for_each is not empty set",
			config: `
resource "aws_instance" "main" {
  for_each = toset(["foo", "bar"])
}
module "aws_instance" {
  for_each = toset(["foo", "bar"])
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
		{
			name: "for_each is empty set",
			config: `
resource "aws_instance" "main" {
  for_each = toset([])
}
module "aws_instance" {
  for_each = toset([])
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{},
		},
		{
			// HINT: Terraform does not allow null as `for_each`
			name: "for_each is null",
			config: `
resource "aws_instance" "main" {
  for_each = null
}
module "aws_instance" {
  for_each = null
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
					{Type: "module", LabelNames: []string{"name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
					{Type: "module", Labels: []string{"aws_instance"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			if err := fs.WriteFile("main.tf", []byte(test.config), os.ModePerm); err != nil {
				t.Fatal(err)
			}

			parser := NewParser(fs)
			mod, diags := parser.LoadConfigDir(".")
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
				cmpopts.IgnoreFields(hcl.Range{}, "Start", "End", "Filename"),
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

func Test_schemaWithDynamic(t *testing.T) {
	tests := []struct {
		name string
		in   *hclext.BodySchema
		want *hclext.BodySchema
	}{
		{
			name: "empty schema",
			in:   &hclext.BodySchema{},
			want: &hclext.BodySchema{},
		},
		{
			name: "attribute schemas",
			in: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
			want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
		},
		{
			name: "block schemas",
			in: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type: "toplevel",
						Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
					},
				},
			},
			want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
				Blocks: []hclext.BlockSchema{
					{
						Type: "toplevel",
						Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "bar"}}},
					},
					{
						Type:       "dynamic",
						LabelNames: []string{"type"},
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
			name: "nested block schemas",
			in: &hclext.BodySchema{
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
			want: &hclext.BodySchema{
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
									LabelNames: []string{"type"},
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
						LabelNames: []string{"type"},
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
												LabelNames: []string{"type"},
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
		t.Run(test.name, func(t *testing.T) {
			got := schemaWithDynamic(test.in)

			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_resolveDynamicBlocks(t *testing.T) {
	tests := []struct {
		name string
		in   *hclext.BodyContent
		want *hclext.BodyContent
	}{
		{
			name: "empty body",
			in:   &hclext.BodyContent{},
			want: &hclext.BodyContent{},
		},
		{
			name: "only attributes",
			in: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
			},
		},
		{
			name: "regular blocks",
			in: &hclext.BodyContent{
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
			want: &hclext.BodyContent{
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
			name: "dynamic blocks",
			in: &hclext.BodyContent{
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
			want: &hclext.BodyContent{
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
			name: "dynamic nested blocks in regular blocks",
			in: &hclext.BodyContent{
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
			want: &hclext.BodyContent{
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
			name: "dynamic nested blocks in dynamic blocks",
			in: &hclext.BodyContent{
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
			want: &hclext.BodyContent{
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
		t.Run(test.name, func(t *testing.T) {
			got := resolveDynamicBlocks(test.in)

			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}
