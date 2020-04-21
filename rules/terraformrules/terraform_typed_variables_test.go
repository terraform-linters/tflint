package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformTypedVariablesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "no type",
			Content: `
variable "no_type" {
  default = "default"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformTypedVariablesRule(),
					Message: "`no_type` variable has no type",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: "complex type",
			Content: `
variable "no_type2" {
  type = list(object({
    internal = number
    external = number
    protocol = string
  }))
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "with type",
			Content: `
variable "with_type" {
  type = string
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformTypedVariablesRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"variables.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
