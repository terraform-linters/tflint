package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsRdsClusterEndpointTagsRule checks whether the resource is tagged correctly
type AwsRdsClusterEndpointTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsRdsClusterEndpointTagsRule returns new tags rule with default attributes
func NewAwsRdsClusterEndpointTagsRule() *AwsRdsClusterEndpointTagsRule {
	return &AwsRdsClusterEndpointTagsRule{
		resourceType:  "aws_rds_cluster_endpoint",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsRdsClusterEndpointTagsRule) Name() string {
	return "aws_resource_tags_aws_rds_cluster_endpoint"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRdsClusterEndpointTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsRdsClusterEndpointTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsRdsClusterEndpointTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsRdsClusterEndpointTagsRule) Check(runner *tflint.Runner) error {
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var tags map[string]string
		err := runner.EvaluateExpr(attribute.Expr, &tags)

		return runner.EnsureNoError(err, func() error {
			configTags := runner.GetConfigTags()
			tagKeys := []string{}
			hash := make(map[string]bool)
			for k, _ := range tags {
				tagKeys = append(tagKeys, k)
				hash[k] = true
			}
			var found []string
			for _, tag := range configTags {
				if _, ok := hash[tag]; ok {
					found = append(found, tag)
				}
			}
			if len(found) != len(configTags) {
				runner.EmitIssue(r, fmt.Sprintf("Wanted tags: %v, found: %v\n", configTags, tags), attribute.Expr.Range() )
			}
			return nil
		})
	})
}
