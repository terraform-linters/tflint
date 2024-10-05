package terraform

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

func TestRebuild(t *testing.T) {
	tests := []struct {
		name    string
		module  *Module
		sources map[string][]byte
		want    *Module
	}{
		{
			name: "HCL native files",
			module: &Module{
				SourceDir: ".",
				Variables: map[string]*Variable{"foo": {Name: "foo"}},
				primaries: map[string]*hcl.File{
					"main.tf": {Bytes: []byte(`variable "foo" { default = 1 }`), Body: hcl.EmptyBody()},
				},
				overrides: map[string]*hcl.File{
					"main_override.tf": {Bytes: []byte(`variable "foo" { default = 2 }`), Body: hcl.EmptyBody()},
					"override.tf":      {Bytes: []byte(`variable "foo" { default = 3 }`), Body: hcl.EmptyBody()},
				},
				Sources: map[string][]byte{
					"main.tf":          []byte(`variable "foo" { default = 1 }`),
					"main_override.tf": []byte(`variable "foo" { default = 2 }`),
					"override.tf":      []byte(`variable "foo" { default = 3 }`),
				},
				Files: map[string]*hcl.File{
					"main.tf":          {Bytes: []byte(`variable "foo" { default = 1 }`), Body: hcl.EmptyBody()},
					"main_override.tf": {Bytes: []byte(`variable "foo" { default = 2 }`), Body: hcl.EmptyBody()},
					"override.tf":      {Bytes: []byte(`variable "foo" { default = 3 }`), Body: hcl.EmptyBody()},
				},
			},
			sources: map[string][]byte{
				"main.tf": []byte(`
variable "foo" { default = 1 }
variable "bar" { default = "bar" }
`),
				"main_override.tf": []byte(`
variable "foo" { default = 2 }
variable "bar" { default = "baz" }
`),
			},
			want: &Module{
				SourceDir: ".",
				Variables: map[string]*Variable{"foo": {Name: "foo"}, "bar": {Name: "bar"}},
				primaries: map[string]*hcl.File{
					"main.tf": {
						Bytes: []byte(`
variable "foo" { default = 1 }
variable "bar" { default = "bar" }
`),
						Body: hcl.EmptyBody(),
					},
				},
				overrides: map[string]*hcl.File{
					"main_override.tf": {
						Bytes: []byte(`
variable "foo" { default = 2 }
variable "bar" { default = "baz" }
`),
						Body: hcl.EmptyBody(),
					},
					"override.tf": {Bytes: []byte(`variable "foo" { default = 3 }`), Body: hcl.EmptyBody()},
				},
				Sources: map[string][]byte{
					"main.tf": []byte(`
variable "foo" { default = 1 }
variable "bar" { default = "bar" }
`),
					"main_override.tf": []byte(`
variable "foo" { default = 2 }
variable "bar" { default = "baz" }
`),
					"override.tf": []byte(`variable "foo" { default = 3 }`),
				},
				Files: map[string]*hcl.File{
					"main.tf": {
						Bytes: []byte(`
variable "foo" { default = 1 }
variable "bar" { default = "bar" }
`),
						Body: hcl.EmptyBody(),
					},
					"main_override.tf": {
						Bytes: []byte(`
variable "foo" { default = 2 }
variable "bar" { default = "baz" }
`),
						Body: hcl.EmptyBody(),
					},
					"override.tf": {Bytes: []byte(`variable "foo" { default = 3 }`), Body: hcl.EmptyBody()},
				},
			},
		},
		{
			name: "HCL JSON files",
			module: &Module{
				SourceDir: ".",
				Variables: map[string]*Variable{"foo": {Name: "foo"}},
				primaries: map[string]*hcl.File{
					"main.tf.json": {Bytes: []byte(`{"variable": {"foo": {"default": 1}}}`), Body: hcl.EmptyBody()},
				},
				overrides: map[string]*hcl.File{
					"main_override.tf.json": {Bytes: []byte(`{"variable": {"foo": {"default": 2}}}`), Body: hcl.EmptyBody()},
					"override.tf.json":      {Bytes: []byte(`{"variable": {"foo": {"default": 3}}}`), Body: hcl.EmptyBody()},
				},
				Sources: map[string][]byte{
					"main.tf.json":          []byte(`{"variable": {"foo": {"default": 1}}}`),
					"main_override.tf.json": []byte(`{"variable": {"foo": {"default": 2}}}`),
					"override.tf.json":      []byte(`{"variable": {"foo": {"default": 3}}}`),
				},
				Files: map[string]*hcl.File{
					"main.tf.json":          {Bytes: []byte(`{"variable": {"foo": {"default": 1}}}`), Body: hcl.EmptyBody()},
					"main_override.tf.json": {Bytes: []byte(`{"variable": {"foo": {"default": 2}}}`), Body: hcl.EmptyBody()},
					"override.tf.json":      {Bytes: []byte(`{"variable": {"foo": {"default": 3}}}`), Body: hcl.EmptyBody()},
				},
			},
			sources: map[string][]byte{
				"main.tf.json":          []byte(`{"variable": {"foo": {"default": 1}, "bar": {"default": "bar"}}}`),
				"main_override.tf.json": []byte(`{"variable": {"foo": {"default": 2}, "bar": {"default": "baz"}}}`),
			},
			want: &Module{
				SourceDir: ".",
				Variables: map[string]*Variable{"foo": {Name: "foo"}, "bar": {Name: "bar"}},
				primaries: map[string]*hcl.File{
					"main.tf.json": {Bytes: []byte(`{"variable": {"foo": {"default": 1}, "bar": {"default": "bar"}}}`), Body: hcl.EmptyBody()},
				},
				overrides: map[string]*hcl.File{
					"main_override.tf.json": {Bytes: []byte(`{"variable": {"foo": {"default": 2}, "bar": {"default": "baz"}}}`), Body: hcl.EmptyBody()},
					"override.tf.json":      {Bytes: []byte(`{"variable": {"foo": {"default": 3}}}`), Body: hcl.EmptyBody()},
				},
				Sources: map[string][]byte{
					"main.tf.json":          []byte(`{"variable": {"foo": {"default": 1}, "bar": {"default": "bar"}}}`),
					"main_override.tf.json": []byte(`{"variable": {"foo": {"default": 2}, "bar": {"default": "baz"}}}`),
					"override.tf.json":      []byte(`{"variable": {"foo": {"default": 3}}}`),
				},
				Files: map[string]*hcl.File{
					"main.tf.json":          {Bytes: []byte(`{"variable": {"foo": {"default": 1}, "bar": {"default": "bar"}}}`), Body: hcl.EmptyBody()},
					"main_override.tf.json": {Bytes: []byte(`{"variable": {"foo": {"default": 2}, "bar": {"default": "baz"}}}`), Body: hcl.EmptyBody()},
					"override.tf.json":      {Bytes: []byte(`{"variable": {"foo": {"default": 3}}}`), Body: hcl.EmptyBody()},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diags := test.module.Rebuild(test.sources)
			if diags.HasErrors() {
				t.Fatalf("unexpected error: %s", diags.Error())
			}

			opt := cmp.Comparer(func(x, y *hcl.File) bool {
				return bytes.Equal(x.Bytes, y.Bytes)
			})

			if diff := cmp.Diff(test.want.Sources, test.module.Sources); diff != "" {
				t.Errorf("sources mismatch:\n%s", diff)
			}
			if diff := cmp.Diff(test.want.Files, test.module.Files, opt); diff != "" {
				t.Errorf("files mismatch:\n%s", diff)
			}
			if diff := cmp.Diff(test.want.primaries, test.module.primaries, opt); diff != "" {
				t.Errorf("primaries mismatch:\n%s", diff)
			}
			if diff := cmp.Diff(test.want.overrides, test.module.overrides, opt); diff != "" {
				t.Errorf("overrides mismatch:\n%s", diff)
			}
			if len(test.want.Variables) != len(test.module.Variables) {
				t.Errorf("variables count mismatch: want %d, got %d", len(test.want.Variables), len(test.module.Variables))
			}
		})
	}
}

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
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main1.tf", Start: hcl.Pos{Line: 3}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main1.tf", Start: hcl.Pos{Line: 2}},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main2.tf", Start: hcl.Pos{Line: 3}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main2.tf", Start: hcl.Pos{Line: 2}},
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
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main_override.tf", Start: hcl.Pos{Line: 3}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2}},
					},
				},
			},
		},
		{
			name: "overrides by multiple files/blocks",
			files: map[string]string{
				"main.tf": `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`,
				"main1_override.tf": `
resource "aws_instance" "foo" {
  instance_type = "m5.2xlarge"
}`,
				"main2_override.tf": `
resource "aws_instance" "foo" {
  instance_type = "m5.4xlarge"
}
resource "aws_instance" "foo" {
  instance_type = "m5.8xlarge"
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
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main2_override.tf", Start: hcl.Pos{Line: 6}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2}},
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
								"foo": &hclext.Attribute{Name: "foo", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3}}},
								"bar": &hclext.Attribute{Name: "bar", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4}}},
							},
							Blocks: hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2}},
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
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 8}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 6}},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 8}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 6}},
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
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 8}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 6}},
					},
					{
						Type:   "module",
						Labels: []string{"bar"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{"instance_type": &hclext.Attribute{Name: "instance_type", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 8}}}},
							Blocks:     hclext.Blocks{},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 6}},
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
										Attributes: hclext.Attributes{"volume_size": &hclext.Attribute{Name: "volume_size", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4}}}},
										Blocks:     hclext.Blocks{},
									},
									DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3}},
								},
							},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2}},
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
										Attributes: hclext.Attributes{"volume_size": &hclext.Attribute{Name: "volume_size", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 11}}}},
										Blocks:     hclext.Blocks{},
									},
									DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 8}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{"volume_size": &hclext.Attribute{Name: "volume_size", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 11}}}},
										Blocks:     hclext.Blocks{},
									},
									DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 8}},
								},
							},
						},
						DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 7}},
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
				cmpopts.IgnoreFields(hcl.Range{}, "End"),
				cmpopts.IgnoreFields(hcl.Pos{}, "Column", "Byte"),
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
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
					},
				},
			},
			Overrides: hclext.Blocks{},
			Want: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
					},
				},
			},
		},
		{
			Name: "no override because resources are difference",
			Primaries: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"baz", "qux"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo2"}},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
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
					Type:   "resource",
					Labels: []string{"foo", "bar"},
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
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "bar"},
							"baz": &hclext.Attribute{Name: "baz"},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "bar"},
							"bar": &hclext.Attribute{Name: "bar"},
							"baz": &hclext.Attribute{Name: "baz"},
						},
					},
				},
			},
		},
		{
			Name: "override nested blocks",
			Primaries: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
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
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "bar"}},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "qux"},
										"bar": &hclext.Attribute{Name: "bar"},
									},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "bar"}},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										// The contents of nested configuration blocks are not merged.
										"baz": &hclext.Attribute{Name: "qux"},
										"bar": &hclext.Attribute{Name: "bar"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "override multiple nested blocks",
			Primaries: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"foo": &hclext.Attribute{Name: "foo"},
									},
								},
							},
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"bar": &hclext.Attribute{Name: "bar"},
									},
								},
							},
							{
								Type: "other_nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
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
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "baz"},
									},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							// Any block types that do not appear in the override block remain from the original block.
							{
								Type: "other_nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"qux": &hclext.Attribute{Name: "qux"},
									},
								},
							},
							// override block replace all blocks of the same type in the original block.
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "baz"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "override lifecycle/provisioner/connection",
			Primaries: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "lifecycle",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"create_before_destroy": &hclext.Attribute{Name: "create_before_destroy"}, "prevent_destroy": &hclext.Attribute{Name: "prevent_destroy"}},
								},
							},
							{
								Type:   "provisioner",
								Labels: []string{"local-exec"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"command": &hclext.Attribute{Name: "command"}},
								},
							},
							{
								Type:   "provisioner",
								Labels: []string{"file"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"content": &hclext.Attribute{Name: "content"}},
								},
							},
							{
								Type: "connection",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"type": &hclext.Attribute{Name: "type"}},
								},
							},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "lifecycle",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"ignore_changes": &hclext.Attribute{Name: "ignore_changes"}, "create_before_destroy": &hclext.Attribute{Name: "create_before_destroy2"}},
								},
							},
							{
								Type:   "provisioner",
								Labels: []string{"remote-exec"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"inline": &hclext.Attribute{Name: "inline"}},
								},
							},
							{
								Type: "connection",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"user": &hclext.Attribute{Name: "user"}},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							// the contents of any lifecycle nested block are merged on an argument-by-argument basis.
							{
								Type: "lifecycle",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"create_before_destroy": &hclext.Attribute{Name: "create_before_destroy2"}, "prevent_destroy": &hclext.Attribute{Name: "prevent_destroy"}, "ignore_changes": &hclext.Attribute{Name: "ignore_changes"}},
								},
							},
							// If an overriding resource block contains one or more provisioner blocks then any provisioner blocks in the original block are ignored.
							{
								Type:   "provisioner",
								Labels: []string{"remote-exec"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"inline": &hclext.Attribute{Name: "inline"}},
								},
							},
							// If an overriding resource block contains a connection block then it completely overrides any connection block present in the original block.
							{
								Type: "connection",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"user": &hclext.Attribute{Name: "user"}},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "override data sources",
			Primaries: hclext.Blocks{
				{
					Type:   "data",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "foo"},
							"bar": &hclext.Attribute{Name: "bar"},
						},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"foo": &hclext.Attribute{Name: "foo"},
									},
								},
							},
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"bar": &hclext.Attribute{Name: "bar"},
									},
								},
							},
							{
								Type: "other_nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
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
					Type:   "data",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "bar"},
							"baz": &hclext.Attribute{Name: "baz"},
						},
						Blocks: hclext.Blocks{
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "baz"},
									},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type:   "data",
					Labels: []string{"foo", "bar"},
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{
							"foo": &hclext.Attribute{Name: "bar"},
							"bar": &hclext.Attribute{Name: "bar"},
							"baz": &hclext.Attribute{Name: "baz"},
						},
						Blocks: hclext.Blocks{
							{
								Type: "other_nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"qux": &hclext.Attribute{Name: "qux"},
									},
								},
							},
							{
								Type: "nested",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"baz": &hclext.Attribute{Name: "baz"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "override locals",
			Primaries: hclext.Blocks{
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}, "bar": &hclext.Attribute{Name: "bar"}},
					},
				},
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"baz": &hclext.Attribute{Name: "baz"}, "qux": &hclext.Attribute{Name: "qux"}},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"baz": &hclext.Attribute{Name: "baz2"}, "foo2": &hclext.Attribute{Name: "foo2"}},
					},
				},
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"bar": &hclext.Attribute{Name: "bar2"}},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo": &hclext.Attribute{Name: "foo"}, "bar": &hclext.Attribute{Name: "bar2"}},
					},
				},
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"baz": &hclext.Attribute{Name: "baz2"}, "qux": &hclext.Attribute{Name: "qux"}},
					},
				},
				// Locals not present in the primaries are added.
				{
					Type: "locals",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"foo2": &hclext.Attribute{Name: "foo2"}},
					},
				},
			},
		},
		{
			Name: "override multiple required_version",
			Primaries: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"required_version": &hclext.Attribute{Name: "required_version1"}},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"required_version": &hclext.Attribute{Name: "required_version2"}},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"required_version": &hclext.Attribute{Name: "required_version3"}},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"required_version": &hclext.Attribute{Name: "required_version4"}},
					},
				},
			},
			Want: hclext.Blocks{
				// When overriding attributes, the last element in override takes precedence,
				// so all attributes of primaries are overridden by required_version4.
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"required_version": &hclext.Attribute{Name: "required_version4"}},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Attributes: hclext.Attributes{"required_version": &hclext.Attribute{Name: "required_version4"}},
					},
				},
			},
		},
		{
			Name: "override required_providers",
			Primaries: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"aws":    &hclext.Attribute{Name: "aws"},
										"google": &hclext.Attribute{Name: "google"},
									},
								},
							},
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"azurerm": &hclext.Attribute{Name: "azurerm"},
									},
								},
							},
						},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"google-beta": &hclext.Attribute{Name: "google-beta"},
									},
								},
							},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"aws":     &hclext.Attribute{Name: "aws2"},
										"azurerm": &hclext.Attribute{Name: "azurerm2"},
									},
								},
							},
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"assert": &hclext.Attribute{Name: "assert"},
									},
								},
							},
						},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"google": &hclext.Attribute{Name: "google2"},
										"time":   &hclext.Attribute{Name: "time"},
									},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"aws":    &hclext.Attribute{Name: "aws2"},
										"google": &hclext.Attribute{Name: "google2"},
									},
								},
							},
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"azurerm": &hclext.Attribute{Name: "azurerm2"},
									},
								},
							},
						},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"google-beta": &hclext.Attribute{Name: "google-beta"},
									},
								},
							},
						},
					},
				},
				// Blocks not present in the primaries are added.
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"assert": &hclext.Attribute{Name: "assert"},
									},
								},
							},
						},
					},
				},
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "required_providers",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{
										"time": &hclext.Attribute{Name: "time"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "override backend",
			Primaries: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type:   "backend",
								Labels: []string{"local"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"path": &hclext.Attribute{Name: "path"}},
								},
							},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type:   "backend",
								Labels: []string{"remote"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"host": &hclext.Attribute{Name: "host"}},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type:   "backend",
								Labels: []string{"remote"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"host": &hclext.Attribute{Name: "host"}},
								},
							},
						},
					},
				},
			},
		},
		{
			// The presence of a block defining a backend (either cloud or backend) in an override file
			// always takes precedence over a block defining a backend in the original configuration
			Name: "override backend by cloud",
			Primaries: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type:   "backend",
								Labels: []string{"local"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"path": &hclext.Attribute{Name: "path"}},
								},
							},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "cloud",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"organization": &hclext.Attribute{Name: "organization"}},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "cloud",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"organization": &hclext.Attribute{Name: "organization"}},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "override cloud by backend",
			Primaries: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type: "cloud",
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"organization": &hclext.Attribute{Name: "organization"}},
								},
							},
						},
					},
				},
			},
			Overrides: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type:   "backend",
								Labels: []string{"remote"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"host": &hclext.Attribute{Name: "host"}},
								},
							},
						},
					},
				},
			},
			Want: hclext.Blocks{
				{
					Type: "terraform",
					Body: &hclext.BodyContent{
						Blocks: hclext.Blocks{
							{
								Type:   "backend",
								Labels: []string{"remote"},
								Body: &hclext.BodyContent{
									Attributes: hclext.Attributes{"host": &hclext.Attribute{Name: "host"}},
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
