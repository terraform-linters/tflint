package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformUnusedDeclarationsRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		JSON     bool
		Expected tflint.Issues
	}{
		{
			Name: "unused variable",
			Content: `
variable "not_used" {}

variable "used" {}
output "u" { value = var.used }
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformUnusedDeclarationsRule(),
					Message: `variable "not_used" is declared but not used`,
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: "unused data source",
			Content: `
data "null_data_source" "not_used" {}

data "null_data_source" "used" {}
output "u" { value = data.null_data_source.used }
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformUnusedDeclarationsRule(),
					Message: `data "null_data_source" "not_used" is declared but not used`,
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 35},
					},
				},
			},
		},
		{
			Name: "unused local source",
			Content: `
locals {
	not_used = ""
	used = ""
}

output "u" { value = local.used }
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformUnusedDeclarationsRule(),
					Message: `local.not_used is declared but not used`,
					Range: hcl.Range{
						Filename: "config.tf",
						Start:    hcl.Pos{Line: 3, Column: 2},
						End:      hcl.Pos{Line: 3, Column: 15},
					},
				},
			},
		},
		{
			Name: "variable used in resource",
			Content: `
variable "used" {}
resource "null_resource" "n" {
	triggers = {
		u = var.used
	}
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "variable used in module",
			Content: `
variable "used" {}
module "m" {
	source = "./module"
	u = var.used
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "variable used in module",
			Content: `
variable "used" {}
module "m" {
	source = "./module"
	u = var.used
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "local used in module",
			Content: `
locals { used = "used" }
module "m" {
	source = "./module"
	u = local.used
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "variable used in provider",
			Content: `
variable "aws_region" {}
provider "aws" {
	region = var.aws_region
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "meta-arguments",
			Content: `
variable "used" {}
resource "null_resource" "n" {
  triggers = {
    u = var.used
	}
  
  lifecycle {
    ignore_changes = [triggers]
  }

  providers = {
    null = null
  }

  depends_on = [aws_instance.foo]
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "additional traversal",
			Content: `
variable "v" {
	type = object({ foo = string })
}
output "v" {
	value = var.v.foo
}

data "terraform_remote_state" "d" {}
output "d" {
	value = data.terraform_remote_state.d.outputs.foo
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "json",
			JSON: true,
			Content: `
{
  "resource": {
    "foo": {
      "bar": {
        "nested": [{
          "${var.again}": []
        }]
      }
    }
	},
  "variable": {
    "again": {}
  }
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformUnusedDeclarationsRule()

	for _, tc := range cases {
		filename := "config.tf"
		if tc.JSON {
			filename += ".json"
		}

		t.Run(tc.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{filename: tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			tflint.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
