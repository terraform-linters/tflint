package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsEc2CapacityReservationTagsRule checks whether the resource is tagged correctly
type AwsEc2CapacityReservationTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsEc2CapacityReservationTagsRule returns new tags rule with default attributes
func NewAwsEc2CapacityReservationTagsRule() *AwsEc2CapacityReservationTagsRule {
	return &AwsEc2CapacityReservationTagsRule{
		resourceType:  "aws_ec2_capacity_reservation",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsEc2CapacityReservationTagsRule) Name() string {
	return "aws_resource_tags_aws_ec2_capacity_reservation"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsEc2CapacityReservationTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsEc2CapacityReservationTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsEc2CapacityReservationTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsEc2CapacityReservationTagsRule) Check(runner *tflint.Runner) error {
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
