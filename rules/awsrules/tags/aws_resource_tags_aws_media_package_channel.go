package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsMediaPackageChannelTagsRule checks whether the resource is tagged correctly
type AwsMediaPackageChannelTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsMediaPackageChannelTagsRule returns new tags rule with default attributes
func NewAwsMediaPackageChannelTagsRule() *AwsMediaPackageChannelTagsRule {
	return &AwsMediaPackageChannelTagsRule{
		resourceType:  "aws_media_package_channel",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsMediaPackageChannelTagsRule) Name() string {
	return "aws_resource_tags_aws_media_package_channel"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsMediaPackageChannelTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsMediaPackageChannelTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsMediaPackageChannelTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsMediaPackageChannelTagsRule) Check(runner *tflint.Runner) error {
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
