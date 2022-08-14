package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformDocumentedVariablesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "no description",
			Content: `
variable "no_description" {
  default = "default"
}`,
			Expected: helper.Issues{
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
			Expected: helper.Issues{
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
			Expected: helper.Issues{},
		},
	}

	rule := NewTerraformDocumentedVariablesRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"variables.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
