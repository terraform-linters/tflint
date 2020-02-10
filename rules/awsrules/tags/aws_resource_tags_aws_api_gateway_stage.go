package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsAPIGatewayStageTagsRule checks whether the resource is tagged correctly
type AwsAPIGatewayStageTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsAPIGatewayStageTagsRule returns new tags rule with default attributes
func NewAwsAPIGatewayStageTagsRule() *AwsAPIGatewayStageTagsRule {
	return &AwsAPIGatewayStageTagsRule{
		resourceType:  "aws_api_gateway_stage",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsAPIGatewayStageTagsRule) Name() string {
	return "aws_resource_tags_aws_api_gateway_stage"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsAPIGatewayStageTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsAPIGatewayStageTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsAPIGatewayStageTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsAPIGatewayStageTagsRule) Check(runner *tflint.Runner) error {
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
