package rules

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsS3BucketExampleLifecycleRuleRule checks whether ...
type AwsS3BucketExampleLifecycleRuleRule struct{}

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
func (r *AwsS3BucketExampleLifecycleRuleRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsS3BucketExampleLifecycleRuleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsS3BucketExampleLifecycleRuleRule) Check(runner tflint.Runner) error {
	return runner.WalkResourceBlocks("aws_s3_bucket", "lifecycle_rule", func(block *hcl.Block) error {
		if err := runner.EmitIssue(r, "`lifecycle_rule` block found", block.DefRange); err != nil {
			return err
		}

		content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "enabled"},
			},
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "transition"},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		if attr, exists := content.Attributes["enabled"]; exists {
			if err := runner.EmitIssueOnExpr(r, "`enabled` attribute found", attr.Expr); err != nil {
				return err
			}
		}

		for _, block := range content.Blocks {
			if err := runner.EmitIssue(r, "`transition` block found", block.DefRange); err != nil {
				return err
			}
		}

		return nil
	})
}
