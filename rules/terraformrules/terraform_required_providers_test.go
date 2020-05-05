package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformRequiredProvidersRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "no version",
			Content: `
provider "template" {}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Provider "template" should have a version constraint in required_providers`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "required_providers set",
			Content: `
terraform {
	required_providers {
		template = "~> 2"
	}
}

provider "template" {} 
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "provider alias",
			Content: `
provider "template" {
	version = "~> 2"
}

provider "template" {
	alias = "b"
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Provider "template" should have a version constraint in required_providers (template.b)`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   6,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "version set",
			Content: `
provider "template" {
	version = "~> 2"
} 
`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformRequiredProvidersRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"module.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
