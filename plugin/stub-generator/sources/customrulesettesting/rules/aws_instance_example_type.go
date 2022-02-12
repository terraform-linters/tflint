package rules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/plugin/stub-generator/sources/customrulesettesting/custom"
)

// AwsInstanceExampleTypeRule checks whether ...
type AwsInstanceExampleTypeRule struct{}

// NewAwsInstanceExampleTypeRule returns a new rule
func NewAwsInstanceExampleTypeRule() *AwsInstanceExampleTypeRule {
	return &AwsInstanceExampleTypeRule{}
}

// Name returns the rule name
func (r *AwsInstanceExampleTypeRule) Name() string {
	return "aws_instance_example_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceExampleTypeRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsInstanceExampleTypeRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceExampleTypeRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsInstanceExampleTypeRule) Check(raw tflint.Runner) error {
	runner := raw.(*custom.Runner)

	return runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
		return runner.EmitIssueOnExpr(
			r,
			fmt.Sprintf(
				"deep_check=%t, token=%s, zone=%s, annotation=%s",
				runner.CustomConfig.DeepCheck,
				runner.CustomConfig.Auth.Token,
				runner.CustomConfig.Zone,
				runner.CustomConfig.Annotation,
			),
			attribute.Expr,
		)
	})
}
