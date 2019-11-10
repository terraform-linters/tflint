package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsElastiCacheClusterPreviousType(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "cache.t1.micro is previous type",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    node_type = "cache.t1.micro"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsElastiCacheClusterPreviousTypeRule(),
					Message: "\"cache.t1.micro\" is previous generation node type.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 17},
						End:      hcl.Pos{Line: 3, Column: 33},
					},
				},
			},
		},
		{
			Name: "cache.t2.micro is not previous type",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    node_type = "cache.t2.micro"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsElastiCacheClusterPreviousTypeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
