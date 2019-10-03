package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsElastiCacheClusterInvalidType(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "t2.micro is invalid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    node_type = "t2.micro"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsElastiCacheClusterInvalidTypeRule(),
					Message: "\"t2.micro\" is invalid node type.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 17},
						End:      hcl.Pos{Line: 3, Column: 27},
					},
				},
			},
		},
		{
			Name: "cache.t2.micro is valid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    node_type = "cache.t2.micro"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsElastiCacheClusterInvalidTypeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
