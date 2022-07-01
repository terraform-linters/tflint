package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformEmptyListEqualityRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "comparing with [] is not recommended",
			Content: `
resource "aws_db_instance" "mysql" {
	count = [] == [] ? 0 : 1
    instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformEmptyListEqualityRule(),
					Message: "Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 10},
						End:      hcl.Pos{Line: 3, Column: 18},
					},
				},
				{
					Rule:    NewTerraformEmptyListEqualityRule(),
					Message: "Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 10},
						End:      hcl.Pos{Line: 3, Column: 18},
					},
				},
			},
		},
		{
			Name: "comparing with [] is not recommended (mixed with other conditions)",
			Content: `
resource "aws_db_instance" "mysql" {
	count = true == true || false != true && (false == false || [] == []) ? 1 : 0
	instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformEmptyListEqualityRule(),
					Message: "Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 62},
						End:      hcl.Pos{Line: 3, Column: 70},
					},
				},
				{
					Rule:    NewTerraformEmptyListEqualityRule(),
					Message: "Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 62},
						End:      hcl.Pos{Line: 3, Column: 70},
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
					Rule:    NewTerraformEmptyListEqualityRule(),
					Message: "Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
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

	rule := NewTerraformEmptyListEqualityRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
