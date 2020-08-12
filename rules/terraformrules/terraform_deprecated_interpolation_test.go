package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDeprecatedInterpolationRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "deprecated single interpolation",
			Content: `
resource "null_resource" "a" {
	triggers = "${var.triggers}"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDeprecatedInterpolationRule(),
					Message: "Interpolation-only expressions are deprecated in Terraform v0.12.14",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 3, Column: 13},
						End:      hcl.Pos{Line: 3, Column: 30},
					},
				},
			},
		},
		{
			Name: "deprecated single interpolation in provider block",
			Content: `
provider "null" {
	foo = "${var.triggers["foo"]}"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDeprecatedInterpolationRule(),
					Message: "Interpolation-only expressions are deprecated in Terraform v0.12.14",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 3, Column: 8},
						End:      hcl.Pos{Line: 3, Column: 32},
					},
				},
			},
		},
		{
			Name: "deprecated single interpolation in locals block",
			Content: `
locals {
	foo = "${var.triggers["foo"]}"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDeprecatedInterpolationRule(),
					Message: "Interpolation-only expressions are deprecated in Terraform v0.12.14",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 3, Column: 8},
						End:      hcl.Pos{Line: 3, Column: 32},
					},
				},
			},
		},
		{
			Name: "deprecated single interpolation in nested block",
			Content: `
resource "null_resource" "a" {
	provisioner "local-exec" {
		single = "${var.triggers["greeting"]}"
	}
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDeprecatedInterpolationRule(),
					Message: "Interpolation-only expressions are deprecated in Terraform v0.12.14",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 4, Column: 12},
						End:      hcl.Pos{Line: 4, Column: 41},
					},
				},
			},
		},
		{
			Name: "interpolation as template",
			Content: `
resource "null_resource" "a" {
	triggers = "${var.triggers} "
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "interpolation in array",
			Content: `
resource "null_resource" "a" {
	triggers = ["${var.triggers}"]
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "new interpolation syntax",
			Content: `
resource "null_resource" "a" {
	triggers = var.triggers
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDeprecatedInterpolationRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"config.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
