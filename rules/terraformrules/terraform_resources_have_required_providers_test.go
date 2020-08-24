package terraformrules

import (
	"github.com/hashicorp/hcl/v2"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformResourcesHaveRequiredProviders(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{

		{
			Name: "no resources",
			Content: `
terraform {
	required_providers {
		template = "~> 2"
	}
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "required_providers set",
			Content: `
terraform {
	required_providers {
		template = "~> 2"
	}
}

resource "template_test" "example" {
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "single resource with alias",
			Content: `
provider "template" {
	alias = "b"
}

terraform {
	required_providers {
		b = "~> 2"
	}
}

resource "template_test" "example" {
    provider = template.b
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "no version",
			Content: `
terraform {
 required_providers {
 }
}

resource "template_test" "example" {
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformResourcesHaveRequiredProvidersRule(),
					Message: `Missing version constraint for provider "template" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   7,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   7,
							Column: 35,
						},
					},
				},
			},
		},
		{
			Name: "no alias version",
			Content: `
terraform {
 required_providers {
    template = "~> 2"
 }
}

resource "template_test" "example" {
    provider = template.b
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformResourcesHaveRequiredProvidersRule(),
					Message: `Missing version constraint for provider "b" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   8,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   8,
							Column: 35,
						},
					},
				},
			},
		},
	}

	rule := NewTerraformResourcesHaveRequiredProvidersRule()

	for _, tc := range cases {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{"module.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			tflint.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
