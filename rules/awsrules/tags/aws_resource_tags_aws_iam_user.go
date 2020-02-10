package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsIAMUserTagsRule checks whether the resource is tagged correctly
type AwsIAMUserTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsIAMUserTagsRule returns new tags rule with default attributes
func NewAwsIAMUserTagsRule() *AwsIAMUserTagsRule {
	return &AwsIAMUserTagsRule{
		resourceType:  "aws_iam_user",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsIAMUserTagsRule) Name() string {
	return "aws_resource_tags_aws_iam_user"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsIAMUserTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsIAMUserTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsIAMUserTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsIAMUserTagsRule) Check(runner *tflint.Runner) error {
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
