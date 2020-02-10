package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsStoragegatewayCachedIscsiVolumeTagsRule checks whether the resource is tagged correctly
type AwsStoragegatewayCachedIscsiVolumeTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsStoragegatewayCachedIscsiVolumeTagsRule returns new tags rule with default attributes
func NewAwsStoragegatewayCachedIscsiVolumeTagsRule() *AwsStoragegatewayCachedIscsiVolumeTagsRule {
	return &AwsStoragegatewayCachedIscsiVolumeTagsRule{
		resourceType:  "aws_storagegateway_cached_iscsi_volume",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsStoragegatewayCachedIscsiVolumeTagsRule) Name() string {
	return "aws_resource_tags_aws_storagegateway_cached_iscsi_volume"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsStoragegatewayCachedIscsiVolumeTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsStoragegatewayCachedIscsiVolumeTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsStoragegatewayCachedIscsiVolumeTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsStoragegatewayCachedIscsiVolumeTagsRule) Check(runner *tflint.Runner) error {
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
