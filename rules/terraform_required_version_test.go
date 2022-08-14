package rules

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformRequiredVersionRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "unset",
			Content: `
terraform {}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredVersionRule(),
					Message: "terraform \"required_version\" attribute is required",
				},
			},
		},
		{
			Name: "set",
			Content: `
terraform {
  required_version = "~> 0.12"
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "multiple blocks",
			Content: `
terraform {
	cloud {
		workspaces {
			name = "foo"
		}
	}
}
terraform {
  required_version = "~> 0.12"
}
`,
			Expected: helper.Issues{},
		},
	}

	rule := NewTerraformRequiredVersionRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"module.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
