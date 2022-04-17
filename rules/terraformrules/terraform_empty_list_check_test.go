package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformEmptyListCheckRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "comparing with [] is not recommended",
			Content: `
variable "my_list" {
	type = list(string)
}
resource "aws_db_instance" "mysql" {
	count = var.my_list == [] ? 0 : 1
    instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformEmptyListCheckRule(),
					Message: "List is compared with [] instead of checking if length is 0.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 10},
						End:      hcl.Pos{Line: 6, Column: 27},
					},
				},
			},
		},
		{
			Name: "negatively comparing with [] is not recommended",
			Content: `
variable "my_list" {
	type = list(string)
}
resource "aws_db_instance" "mysql" {
	count = var.my_list != [] ? 1 : 0
    instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformEmptyListCheckRule(),
					Message: "List is compared with [] instead of checking if length is 0.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 10},
						End:      hcl.Pos{Line: 6, Column: 27},
					},
				},
			},
		},
		{
			Name: "checking if length is 0 is recommended",
			Content: `
variable "my_list" {
	type = list(string)
}
resource "aws_db_instance" "mysql" {
	count = length(var.my_list) == 0 ? 1 : 0
	instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformEmptyListCheckRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
