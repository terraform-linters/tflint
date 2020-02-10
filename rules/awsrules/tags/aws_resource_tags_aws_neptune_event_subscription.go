package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsNeptuneEventSubscriptionTagsRule checks whether the resource is tagged correctly
type AwsNeptuneEventSubscriptionTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsNeptuneEventSubscriptionTagsRule returns new tags rule with default attributes
func NewAwsNeptuneEventSubscriptionTagsRule() *AwsNeptuneEventSubscriptionTagsRule {
	return &AwsNeptuneEventSubscriptionTagsRule{
		resourceType:  "aws_neptune_event_subscription",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsNeptuneEventSubscriptionTagsRule) Name() string {
	return "aws_resource_tags_aws_neptune_event_subscription"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsNeptuneEventSubscriptionTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsNeptuneEventSubscriptionTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsNeptuneEventSubscriptionTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsNeptuneEventSubscriptionTagsRule) Check(runner *tflint.Runner) error {
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var resourceTags map[string]string
		err := runner.EvaluateExpr(attribute.Expr, &resourceTags)
		tags := []string{}
		for k := range resourceTags {
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
				sort.Strings(configTags)
				sort.Strings(tags)
				wanted := strings.Join(configTags, ",")
				found := strings.Join(tags, ",")
				runner.EmitIssue(r, fmt.Sprintf("Wanted tags: %v, found: %v\n", wanted, found), attribute.Expr.Range())
			}
			return nil
		})
	})
}
