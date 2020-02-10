package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsSagemakerModelTagsRule checks whether the resource is tagged correctly
type AwsSagemakerModelTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsSagemakerModelTagsRule returns new tags rule with default attributes
func NewAwsSagemakerModelTagsRule() *AwsSagemakerModelTagsRule {
	return &AwsSagemakerModelTagsRule{
		resourceType:  "aws_sagemaker_model",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsSagemakerModelTagsRule) Name() string {
	return "aws_resource_tags_aws_sagemaker_model"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsSagemakerModelTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsSagemakerModelTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsSagemakerModelTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsSagemakerModelTagsRule) Check(runner *tflint.Runner) error {
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
