package awsrules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsSfnActivityTagsRule checks whether the resource is tagged correctly
type AwsSfnActivityTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsSfnActivityTagsRule returns new tags rule with default attributes
func NewAwsSfnActivityTagsRule() *AwsSfnActivityTagsRule {
	return &AwsSfnActivityTagsRule{
		resourceType:  "aws_sfn_activity",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsSfnActivityTagsRule) Name() string {
	return "aws_resource_tags_aws_sfn_activity"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsSfnActivityTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsSfnActivityTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsSfnActivityTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsSfnActivityTagsRule) Check(runner *tflint.Runner) error {
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
