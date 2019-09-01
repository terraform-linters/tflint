package awsrules

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsElastiCacheClusterDefaultParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "default.redis3.2 is default parameter group",
			Content: `
resource "aws_elasticache_cluster" "cache" {
    parameter_group_name = "default.redis3.2"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsElastiCacheClusterDefaultParameterGroupRule(),
					Message: "\"default.redis3.2\" is default parameter group. You cannot edit it.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 46},
					},
				},
			},
		},
		{
			Name: "application3.2 is not default parameter group",
			Content: `
resource "aws_elasticache_cluster" "cache" {
    parameter_group_name = "application3.2"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsElastiCacheClusterDefaultParameterGroupRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
