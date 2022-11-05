package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsS3BucketExampleLifecycleRuleRule checks whether ...
type AwsS3BucketExampleLifecycleRuleRule struct {
	tflint.DefaultRule
}

// NewAwsS3BucketExampleLifecycleRuleRule returns a new rule
func NewAwsS3BucketExampleLifecycleRuleRule() *AwsS3BucketExampleLifecycleRuleRule {
	return &AwsS3BucketExampleLifecycleRuleRule{}
}

// Name returns the rule name
func (r *AwsS3BucketExampleLifecycleRuleRule) Name() string {
	return "aws_s3_bucket_example_lifecycle_rule"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsS3BucketExampleLifecycleRuleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsS3BucketExampleLifecycleRuleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsS3BucketExampleLifecycleRuleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsS3BucketExampleLifecycleRuleRule) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_s3_bucket", &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "lifecycle_rule",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "enabled"},
					},
					Blocks: []hclext.BlockSchema{
						{
							Type: "transition",
							Body: &hclext.BodySchema{
								Attributes: []hclext.AttributeSchema{
									{Name: "days"},
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		for _, lifecycle := range resource.Body.Blocks {
			if err := runner.EmitIssue(r, "`lifecycle_rule` block found", lifecycle.DefRange); err != nil {
				return err
			}

			if attr, exists := lifecycle.Body.Attributes["enabled"]; exists {
				var enabled string
				err := runner.EnsureNoError(runner.EvaluateExpr(attr.Expr, &enabled, nil), func() error {
					return runner.EmitIssue(r, fmt.Sprintf("`enabled` attribute found: %s", enabled), attr.Expr.Range())
				})
				if err != nil {
					return err
				}
			}

			for _, transition := range lifecycle.Body.Blocks {
				if err := runner.EmitIssue(r, "`transition` block found", transition.DefRange); err != nil {
					return err
				}

				if attr, exists := transition.Body.Attributes["days"]; exists {
					var days int
					err := runner.EnsureNoError(runner.EvaluateExpr(attr.Expr, &days, nil), func() error {
						return runner.EmitIssue(r, fmt.Sprintf("`days` attribute found: %d", days), attr.Expr.Range())
					})
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
