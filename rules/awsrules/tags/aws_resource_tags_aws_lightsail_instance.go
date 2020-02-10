package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsLightsailInstanceTagsRule checks whether the resource is tagged correctly
type AwsLightsailInstanceTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsLightsailInstanceTagsRule returns new tags rule with default attributes
func NewAwsLightsailInstanceTagsRule() *AwsLightsailInstanceTagsRule {
	return &AwsLightsailInstanceTagsRule{
		resourceType:  "aws_lightsail_instance",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsLightsailInstanceTagsRule) Name() string {
	return "aws_resource_tags_aws_lightsail_instance"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsLightsailInstanceTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsLightsailInstanceTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsLightsailInstanceTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsLightsailInstanceTagsRule) Check(runner *tflint.Runner) error {
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var resourceTags map[string]string
		err := runner.EvaluateExpr(attribute.Expr, &resourceTags)
		tags := []string{}
		for k, _ := range resourceTags {
			tags = append(tags, k)
		}

		return runner.EnsureNoError(err, func() error {
			configTags := runner.GetConfigTags()
			hash := make(map[string]bool)
			for _, k := range tags {
				hash[k] = true
			}
			var found []string
			for _, tag := range configTags {
				if _, ok := hash[tag]; ok {
					found = append(found, tag)
				}
			}
			if len(found) != len(configTags) {
				wanted := strings.Join(sort.StringSlice(configTags), ",")
				found := strings.Join(sort.StringSlice(tags), ",")
				runner.EmitIssue(r, fmt.Sprintf("Wanted tags: %v, found: %v\n", wanted, found), attribute.Expr.Range())
			}
			return nil
		})
	})
}
