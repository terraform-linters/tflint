package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDocumentedOutputsRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "no description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
}`,
			Expected: tflint.Issues{
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
			Expected: tflint.Issues{
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
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDocumentedOutputsRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"outputs.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
