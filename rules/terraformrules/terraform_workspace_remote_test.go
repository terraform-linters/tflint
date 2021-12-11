package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformWorkspaceRemoteRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "terraform.workspace in resource with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
resource "null_resource" "a" {
	triggers = {
		w = terraform.workspace
	}
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformWorkspaceRemoteRule(),
					Message: "terraform.workspace should not be used with a 'remote' backend",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 6, Column: 13},
						End:      hcl.Pos{Line: 8, Column: 3},
					},
				},
			},
		},
		{
			Name: "terraform.workspace with no backend",
			Content: `
resource "null_resource" "a" {
	triggers = {
		w = terraform.workspace
	}
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "terraform.workspace with non-remote backend",
			Content: `
terraform {
	backend "local" {}
}

resource "null_resource" "a" {
	triggers = {
		w = terraform.workspace
	}
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "terraform.workspace in data source with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
data "null_data_source" "a" {
	inputs = {
		w = terraform.workspace
	}
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformWorkspaceRemoteRule(),
					Message: "terraform.workspace should not be used with a 'remote' backend",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 6, Column: 11},
						End:      hcl.Pos{Line: 8, Column: 3},
					},
				},
			},
		},
		{
			Name: "terraform.workspace in module call with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
module "a" {
	source = "./module"
	w = terraform.workspace
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformWorkspaceRemoteRule(),
					Message: "terraform.workspace should not be used with a 'remote' backend",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 7, Column: 6},
						End:      hcl.Pos{Line: 7, Column: 25},
					},
				},
			},
		},
		{
			Name: "terraform.workspace in provider config with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
provider "aws" {
	assume_role	{
		role_arn = terraform.workspace
	}
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformWorkspaceRemoteRule(),
					Message: "terraform.workspace should not be used with a 'remote' backend",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 7, Column: 14},
						End:      hcl.Pos{Line: 7, Column: 33},
					},
				},
			},
		},
		{
			Name: "terraform.workspace in locals with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
locals {
	w = terraform.workspace
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformWorkspaceRemoteRule(),
					Message: "terraform.workspace should not be used with a 'remote' backend",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 6, Column: 6},
						End:      hcl.Pos{Line: 6, Column: 25},
					},
				},
			},
		},
		{
			Name: "terraform.workspace in output with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
output "o" {
	value = terraform.workspace
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformWorkspaceRemoteRule(),
					Message: "terraform.workspace should not be used with a 'remote' backend",
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 6, Column: 10},
						End:      hcl.Pos{Line: 6, Column: 29},
					},
				},
			},
		},
		{
			Name: "nonmatching expressions with remote backend",
			Content: `
terraform {
	backend "remote" {}
}
locals {
	a = "terraform.workspace"
	b = path.module
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "meta-arguments",
			Content: `
terraform {
	backend "remote" {}
}
resource "aws_instance" "foo" {
  instance_type = "t3.nano"

  lifecycle {
    ignore_changes = [instance_type]
  }

  providers = {
    aws = aws
  }

  depends_on = [aws_instance.bar]
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformWorkspaceRemoteRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{"config.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			tflint.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
