package awsrules

import (
	"fmt"
	"sort"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// TagsRule checks whether the resource is tagged correctly
type TagsRule struct {
	resourceType  string
	attributeName string
	ruleName      string
}

// NewTagsRules returns new rules for all resources that support tags
func NewTagsRules(resourceTypes []string) []*TagsRule {
	tagsRules := []*TagsRule{}
	for _, resourceType := range resourceTypes {
		tagsRules = append(tagsRules, newTagsRule(resourceType))
	}
	return tagsRules
}

// newTagsRule returns new tags rule with default attributes
func newTagsRule(resourceType string) *TagsRule {
	return &TagsRule{
		resourceType:  resourceType,
		attributeName: "tags",
		ruleName:      "aws_resource_tags_" + resourceType,
	}
}

// Name returns the rule name
func (r *TagsRule) Name() string {
	return r.ruleName
}

// Enabled returns whether the rule is enabled by default
func (r *TagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *TagsRule) Check(runner *tflint.Runner) error {
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
