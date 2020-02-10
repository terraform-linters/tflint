package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsWafregionalRuleTagsRule checks whether the resource is tagged correctly
type AwsWafregionalRuleTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsWafregionalRuleTagsRule returns new tags rule with default attributes
func NewAwsWafregionalRuleTagsRule() *AwsWafregionalRuleTagsRule {
	return &AwsWafregionalRuleTagsRule{
		resourceType:  "aws_wafregional_rule",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsWafregionalRuleTagsRule) Name() string {
	return "aws_resource_tags_aws_wafregional_rule"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsWafregionalRuleTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsWafregionalRuleTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsWafregionalRuleTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsWafregionalRuleTagsRule) Check(runner *tflint.Runner) error {
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
