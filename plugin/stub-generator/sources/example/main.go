package main

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: tflint.RuleSet{
			Name:    "example",
			Version: "0.1.0",
			Rules: []tflint.Rule{
				NewAwsInstanceExampleTypeRule(),
			},
		},
	})
}

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
func (r *AwsInstanceExampleTypeRule) Check(runner tflint.Runner) error {
	return runner.WalkResourceAttributes("aws_instance", "instance_type", func(attribute *hcl.Attribute) error {
		var instanceType string
		err := runner.EvaluateExpr(attribute.Expr, &instanceType)

		return runner.EnsureNoError(err, func() error {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("instance type is %s", instanceType),
				attribute.Expr.Range(),
				tflint.Metadata{Expr: attribute.Expr},
			)
		})
	})
}
