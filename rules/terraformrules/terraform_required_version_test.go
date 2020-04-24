package terraformrules

import (
	"github.com/hashicorp/hcl/v2"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformRequiredVersionRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: "no version",
			Content: `
terraform {}
`,
			Config: `
rule "terraform_required_version" {
  enabled = true
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredVersionRule(),
					Message: "terraform \"required_version\" attribute is required",
				},
			},
		},
		{
			Name: "version not match",
			Content: `
terraform {
  required_version = "> 0.12"
}
`,
			Config: `
rule "terraform_required_version" {
  enabled = true
  version = "~> 0.12"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformRequiredVersionRule(),
					Message: "terraform \"required_version\" does not match specified version \"~> 0.12\"",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 30},
					},
				},
			},
		},
		{
			Name: "version matches",
			Content: `
terraform {
  required_version = "~> 0.12"
}
`,
			Config: `
rule "terraform_required_version" {
  enabled = true
  version = "~> 0.12"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "version exists",
			Content: `
terraform {
  required_version = "~> 0.12"
}
`,
			Config: `
rule "terraform_required_version" {
  enabled = true
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformRequiredVersionRule()

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"module.tf": tc.Content}, loadConfigfromTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
