package awsrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsElastiCacheClusterInvalidTypeRule checks whether "aws_elasticache_cluster" has invalid node type.
type AwsElastiCacheClusterInvalidTypeRule struct {
	resourceType  string
	attributeName string
	nodeTypes     map[string]bool
}

// NewAwsElastiCacheClusterInvalidTypeRule returns new rule with default attributes
func NewAwsElastiCacheClusterInvalidTypeRule() *AwsElastiCacheClusterInvalidTypeRule {
	return &AwsElastiCacheClusterInvalidTypeRule{
		resourceType:  "aws_elasticache_cluster",
		attributeName: "node_type",
		nodeTypes: map[string]bool{
			"cache.t2.micro":    true,
			"cache.t2.small":    true,
			"cache.t2.medium":   true,
			"cache.m3.medium":   true,
			"cache.m3.large":    true,
			"cache.m3.xlarge":   true,
			"cache.m3.2xlarge":  true,
			"cache.m4.large":    true,
			"cache.m4.xlarge":   true,
			"cache.m4.2xlarge":  true,
			"cache.m4.4xlarge":  true,
			"cache.m4.10xlarge": true,
			"cache.r3.large":    true,
			"cache.r3.xlarge":   true,
			"cache.r3.2xlarge":  true,
			"cache.r3.4xlarge":  true,
			"cache.r3.8xlarge":  true,
			"cache.r4.large":    true,
			"cache.r4.xlarge":   true,
			"cache.r4.2xlarge":  true,
			"cache.r4.4xlarge":  true,
			"cache.r4.8xlarge":  true,
			"cache.r4.16xlarge": true,
			"cache.m1.small":    true,
			"cache.m1.medium":   true,
			"cache.m1.large":    true,
			"cache.m1.xlarge":   true,
			"cache.m2.xlarge":   true,
			"cache.m2.2xlarge":  true,
			"cache.m2.4xlarge":  true,
			"cache.c1.xlarge":   true,
			"cache.t1.micro":    true,
		},
	}
}

// Name returns the rule name
func (r *AwsElastiCacheClusterInvalidTypeRule) Name() string {
	return "aws_elasticache_cluster_invalid_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsElastiCacheClusterInvalidTypeRule) Enabled() bool {
	return true
}

// Check checks whether "aws_elasticache_cluster" has invalid node type.
func (r *AwsElastiCacheClusterInvalidTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var nodeType string
		err := runner.EvaluateExpr(attribute.Expr, &nodeType)

		return runner.EnsureNoError(err, func() error {
			if !r.nodeTypes[nodeType] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid node type.", nodeType),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_elasticache_cluster_invalid_type.md",
				})
			}
			return nil
		})
	})
}
