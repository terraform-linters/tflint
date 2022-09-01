package rules

import (
	"errors"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsCloudFormationStackErrorRule checks whether ...
type AwsCloudFormationStackErrorRule struct {
	tflint.DefaultRule
}

// NewAwsCloudFormationStackErrorRule returns a new rule
func NewAwsCloudFormationStackErrorRule() *AwsCloudFormationStackErrorRule {
	return &AwsCloudFormationStackErrorRule{}
}

// Name returns the rule name
func (r *AwsCloudFormationStackErrorRule) Name() string {
	return "aws_cloudformation_stack_error"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsCloudFormationStackErrorRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsCloudFormationStackErrorRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsCloudFormationStackErrorRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsCloudFormationStackErrorRule) Check(runner tflint.Runner) error {

	resources, err := runner.GetResourceContent("aws_cloudformation_stack", &hclext.BodySchema{}, nil)
	if err != nil {
		return err
	}

	if len(resources.Blocks) > 0 {
		return errors.New("an error occurred in Check")
	}
	return nil
}
