package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDocumentedVariablesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "no description",
			Content: `
variable "no_description" {
  default = "default"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDocumentedVariablesRule(),
					Message: "`no_description` variable has no description",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 26},
					},
				},
			},
		},
		{
			Name: "empty description",
			Content: `
variable "empty_description" {
  description = ""
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDocumentedVariablesRule(),
					Message: "`empty_description` variable has no description",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 29},
					},
				},
			},
		},
		{
			Name: "with description",
			Content: `
variable "with_description" {
  description = "This is description"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDocumentedVariablesRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"variables.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
