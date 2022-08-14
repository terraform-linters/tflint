package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformDocumentedOutputsRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "no description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformDocumentedOutputsRule(),
					Message: "`endpoint` output has no description",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: "empty description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
  description = ""
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformDocumentedOutputsRule(),
					Message: "`endpoint` output has no description",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: "with description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
  description = "DNS Endpoint"
}`,
			Expected: helper.Issues{},
		},
	}

	rule := NewTerraformDocumentedOutputsRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"outputs.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
