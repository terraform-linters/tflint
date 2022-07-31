package rules

import (
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsS3BucketExampleLifecycleRule checks whether ...
type AwsS3BucketExampleLifecycleRule struct {
	tflint.DefaultRule
}

// NewAwsS3BucketExampleLifecycleRule returns a new rule
func NewAwsS3BucketExampleLifecycleRule() *AwsS3BucketExampleLifecycleRule {
	return &AwsS3BucketExampleLifecycleRule{}
}

// Name returns the rule name
func (r *AwsS3BucketExampleLifecycleRule) Name() string {
	return "aws_s3_bucket_example_lifecycle_rule"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsS3BucketExampleLifecycleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsS3BucketExampleLifecycleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsS3BucketExampleLifecycleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsS3BucketExampleLifecycleRule) Check(runner tflint.Runner) error {
	// This rule is an example to get nested resource attributes.
	resources, err := runner.GetResourceContent("aws_s3_bucket", &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "lifecycle_rule",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "enabled"},
					},
					Blocks: []hclext.BlockSchema{
						{Type: "transition"},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		for _, rule := range resource.Body.Blocks {
			if err := runner.EmitIssue(r, "`lifecycle_rule` block found", rule.DefRange); err != nil {
				return err
			}

			if attr, exists := rule.Body.Attributes["enabled"]; exists {
				if err := runner.EmitIssue(r, "`enabled` attribute found", attr.Expr.Range()); err != nil {
					return err
				}
			}

			for _, transitions := range rule.Body.Blocks {
				if err := runner.EmitIssue(r, "`transition` block found", transitions.DefRange); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
