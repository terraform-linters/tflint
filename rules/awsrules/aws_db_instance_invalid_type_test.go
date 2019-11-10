package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsDBInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "m4.2xlarge is invalid",
			Content: `
resource "aws_db_instance" "mysql" {
    instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidTypeRule(),
					Message: "\"m4.2xlarge\" is invalid instance type.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 22},
						End:      hcl.Pos{Line: 3, Column: 34},
					},
				},
			},
		},
		{
			Name: "db.m4.2xlarge is valid",
			Content: `
resource "aws_db_instance" "mysql" {
    instance_class = "db.m4.2xlarge"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsDBInstanceInvalidTypeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
