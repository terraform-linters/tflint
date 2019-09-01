package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

func Test_TerraformDashInResourceNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in resource name",
			Content: `
resource "aws_eip" "dash-name" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDashInResourceNameRule(),
					Message: "`dash-name` resource name has a dash",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 31},
					},
				},
			},
		},
		{
			Name: "no dash in resource name",
			Content: `
resource "aws_eip" "no_dash_name" {
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDashInResourceNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resources.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
