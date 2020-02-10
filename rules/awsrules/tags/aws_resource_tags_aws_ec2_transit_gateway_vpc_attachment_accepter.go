package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule checks whether the resource is tagged correctly
type AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsEc2TransitGatewayVpcAttachmentAccepterTagsRule returns new tags rule with default attributes
func NewAwsEc2TransitGatewayVpcAttachmentAccepterTagsRule() *AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule {
	return &AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule{
		resourceType:  "aws_ec2_transit_gateway_vpc_attachment_accepter",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule) Name() string {
	return "aws_resource_tags_aws_ec2_transit_gateway_vpc_attachment_accepter"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsEc2TransitGatewayVpcAttachmentAccepterTagsRule) Check(runner *tflint.Runner) error {
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
