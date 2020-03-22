package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDashInDataSourceNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in resource name",
			Content: `
data "aws_eip" "dash-name" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDashInDataSourceNameRule(),
					Message: "`dash-name` data source name has a dash",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 27},
					},
				},
			},
		},
		{
			Name: "no dash in resource name",
			Content: `
data "aws_eip" "no_dash_name" {
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDashInDataSourceNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resources.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
