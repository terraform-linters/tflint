package rules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// AwsInstanceMapEvalExampleRule checks whether ...
type AwsInstanceMapEvalExampleRule struct{}

// NewAwsInstanceMapEvalExampleRule returns a new rule
func NewAwsInstanceMapEvalExampleRule() *AwsInstanceMapEvalExampleRule {
	return &AwsInstanceMapEvalExampleRule{}
}

// Name returns the rule name
func (r *AwsInstanceMapEvalExampleRule) Name() string {
	return "aws_instance_map_eval_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceMapEvalExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsInstanceMapEvalExampleRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceMapEvalExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsInstanceMapEvalExampleRule) Check(runner tflint.Runner) error {
	return runner.WalkResourceAttributes("aws_instance", "tags", func(attribute *hcl.Attribute) error {
		wantType := cty.Map(cty.String)
		tags := map[string]string{}
		err := runner.EvaluateExpr(attribute.Expr, &tags, &wantType)

		return runner.EnsureNoError(err, func() error {
			return runner.EmitIssueOnExpr(
				r,
				fmt.Sprintf("instance tags: %#v", tags),
				attribute.Expr,
			)
		})
	})
}
