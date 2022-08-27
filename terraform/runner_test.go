package terraform

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/zclconf/go-cty/cty"
)

func TestGetModuleCalls(t *testing.T) {
	parseExpr := func(t *testing.T, src string, start hcl.Pos) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "main.tf", start)
		if diags.HasErrors() {
			t.Fatalf("failed to setup test; parse error on `%s`; %s", src, diags)
		}
		return expr
	}

	tests := []struct {
		name    string
		content string
		want    []*ModuleCall
	}{
		{
			name: "local module",
			content: `
module "server" {
  source = "./server"
}`,
			want: []*ModuleCall{
				{
					Name: "server",
					DefRange: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 16},
					},
					Source: "./server",
					SourceAttr: &hclext.Attribute{
						Name: "source",
						Expr: parseExpr(t, `"./server"`, hcl.Pos{Line: 3, Column: 12}),
						Range: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 3, Column: 3},
							End:      hcl.Pos{Line: 3, Column: 22},
						},
						NameRange: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 3, Column: 3},
							End:      hcl.Pos{Line: 3, Column: 9},
						},
					},
				},
			},
		},
		{
			name: "registry module",
			content: `
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.14.2"
}`,
			want: []*ModuleCall{
				{
					Name: "vpc",
					DefRange: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 13},
					},
					Source: "terraform-aws-modules/vpc/aws",
					SourceAttr: &hclext.Attribute{
						Name: "source",
						Expr: parseExpr(t, `"terraform-aws-modules/vpc/aws"`, hcl.Pos{Line: 3, Column: 13}),
						Range: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 3, Column: 3},
							End:      hcl.Pos{Line: 3, Column: 44},
						},
						NameRange: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 3, Column: 3},
							End:      hcl.Pos{Line: 3, Column: 9},
						},
					},
					Version: version.MustConstraints(version.NewConstraint("3.14.2")),
					VersionAttr: &hclext.Attribute{
						Name: "version",
						Expr: parseExpr(t, `"3.14.2"`, hcl.Pos{Line: 4, Column: 13}),
						Range: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 4, Column: 3},
							End:      hcl.Pos{Line: 4, Column: 21},
						},
						NameRange: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 4, Column: 3},
							End:      hcl.Pos{Line: 4, Column: 10},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := NewRunner(helper.TestRunner(t, map[string]string{"main.tf": test.content}))

			got, diags := runner.GetModuleCalls()
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
				cmp.Comparer(func(x, y cty.Value) bool {
					return x.GoString() == y.GoString()
				}),
				cmpopts.IgnoreUnexported(version.Constraint{}),
			}
			if diff := cmp.Diff(got, test.want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestGetLocals(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]*Local
	}{
		{
			name: "locals",
			content: `
locals {
  foo = "bar"
  bar = "baz"
  baz = 1
}`,
			want: map[string]*Local{
				"foo": {Name: "foo", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 3}, End: hcl.Pos{Line: 3, Column: 14}}},
				"bar": {Name: "bar", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 3}, End: hcl.Pos{Line: 4, Column: 14}}},
				"baz": {Name: "baz", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 3}, End: hcl.Pos{Line: 5, Column: 10}}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := NewRunner(helper.TestRunner(t, map[string]string{"main.tf": test.content}))

			got, diags := runner.GetLocals()
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
			}
			if diff := cmp.Diff(got, test.want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestGetProviderRefs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]*ProviderRef
	}{
		{
			name: "resource",
			content: `
resource "google_compute_instance" "main" {
  provider = google.europe
}`,
			want: map[string]*ProviderRef{
				"google": {Name: "google", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 42}}},
			},
		},
		{
			name: "data",
			content: `
data "aws_ami" "main" {
  provider = aws.west
}`,
			want: map[string]*ProviderRef{
				"aws": {Name: "aws", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 22}}},
			},
		},
		{
			name: "provider",
			content: `
provider "google" {
  project = "my-awesome-project"
}`,
			want: map[string]*ProviderRef{
				"google": {Name: "google", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 18}}},
			},
		},
		{
			name: "module",
			content: `
module "server" {
  providers = {
    aws = aws.usw2
  }
}`,
			want: map[string]*ProviderRef{
				"aws": {Name: "aws", DefRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 16}}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := NewRunner(helper.TestRunner(t, map[string]string{"main.tf": test.content}))

			got, diags := runner.GetProviderRefs()
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
			}
			if diff := cmp.Diff(got, test.want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}
