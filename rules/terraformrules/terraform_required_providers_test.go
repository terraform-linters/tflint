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
					Message: `Missing version constraint for provider "template" in "required_providers"`,
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
			Name: "single provider with alias",
			Content: `
provider "template" {
	alias = "b"
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "template" in "required_providers"`,
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
			Name: "version set with alias",
			Content: `
terraform {
  required_providers {
    template = "~> 2"
  }
}

provider "template" {
	version = "~> 2"
} 
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `provider.template: version constraint should be specified via "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   8,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   8,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "version set",
			Content: `
terraform {
  required_providers {
    template = "~> 2"
  }
}

provider "template" {
	alias   = "foo"
	version = "~> 2"
} 
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `provider.template.foo: version constraint should be specified via "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   8,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   8,
							Column: 20,
						},
					},
				},
			},
		},
	}

	rule := NewTerraformRequiredProvidersRule()

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
