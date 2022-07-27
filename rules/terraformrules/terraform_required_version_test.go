package terraformrules

import (
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformRequiredVersionRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "unset",
			Content: `
terraform {}
`,
			Expected: tflint.Issues{
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformRequiredVersionRule()

	for _, tc := range cases {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{"module.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			tflint.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
