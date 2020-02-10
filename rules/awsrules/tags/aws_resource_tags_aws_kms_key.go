package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsKmsKeyTagsRule checks whether the resource is tagged correctly
type AwsKmsKeyTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsKmsKeyTagsRule returns new tags rule with default attributes
func NewAwsKmsKeyTagsRule() *AwsKmsKeyTagsRule {
	return &AwsKmsKeyTagsRule{
		resourceType:  "aws_kms_key",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsKmsKeyTagsRule) Name() string {
	return "aws_resource_tags_aws_kms_key"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsKmsKeyTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsKmsKeyTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsKmsKeyTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsKmsKeyTagsRule) Check(runner *tflint.Runner) error {
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
