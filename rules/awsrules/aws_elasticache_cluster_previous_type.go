package awsrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsElastiCacheClusterPreviousTypeRule checks whether the resource uses previous generation node type
type AwsElastiCacheClusterPreviousTypeRule struct {
	resourceType      string
	attributeName     string
	previousNodeTypes map[string]bool
}

// NewAwsElastiCacheClusterPreviousTypeRule returns new rule with default attributes
func NewAwsElastiCacheClusterPreviousTypeRule() *AwsElastiCacheClusterPreviousTypeRule {
	return &AwsElastiCacheClusterPreviousTypeRule{
		resourceType:  "aws_elasticache_cluster",
		attributeName: "node_type",
		previousNodeTypes: map[string]bool{
			"cache.m1.small":   true,
			"cache.m1.medium":  true,
			"cache.m1.large":   true,
			"cache.m1.xlarge":  true,
			"cache.m2.xlarge":  true,
			"cache.m2.2xlarge": true,
			"cache.m2.4xlarge": true,
			"cache.c1.xlarge":  true,
			"cache.t1.micro":   true,
		},
	}
}

// Name returns the rule name
func (r *AwsElastiCacheClusterPreviousTypeRule) Name() string {
	return "aws_elasticache_cluster_previous_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsElastiCacheClusterPreviousTypeRule) Enabled() bool {
	return true
}

// Check checks whether the resource's `node_type` is included in the list of previous generation node type
func (r *AwsElastiCacheClusterPreviousTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var nodeType string
		err := runner.EvaluateExpr(attribute.Expr, &nodeType)

		return runner.EnsureNoError(err, func() error {
			if r.previousNodeTypes[nodeType] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.WARNING,
					Message:  fmt.Sprintf("\"%s\" is previous generation node type.", nodeType),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_elasticache_cluster_previous_type.md",
				})
			}
			return nil
		})
	})
}
