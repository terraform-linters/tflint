package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsDBInstanceWithDefaultConfigExampleRule checks whether ...
type AwsDBInstanceWithDefaultConfigExampleRule struct {
	tflint.DefaultRule
}

type awsDBInstanceWithDefaultConfigExampleRule struct {
	Name string `hclext:"name,optional"`
}

// NewAwsDBInstanceWithDefaultConfigExampleRule returns a new rule
func NewAwsDBInstanceWithDefaultConfigExampleRule() *AwsDBInstanceWithDefaultConfigExampleRule {
	return &AwsDBInstanceWithDefaultConfigExampleRule{}
}

// Name returns the rule name
func (r *AwsDBInstanceWithDefaultConfigExampleRule) Name() string {
	return "aws_db_instance_with_default_config_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceWithDefaultConfigExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDBInstanceWithDefaultConfigExampleRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *AwsDBInstanceWithDefaultConfigExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsDBInstanceWithDefaultConfigExampleRule) Check(runner tflint.Runner) error {
	config := awsDBInstanceWithDefaultConfigExampleRule{Name: "default"}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	resources, err := runner.GetResourceContent("aws_db_instance", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "name"}},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		attribute, exists := resource.Body.Attributes["name"]
		if !exists {
			continue
		}

		err := runner.EvaluateExpr(attribute.Expr, func(name string) error {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("DB name is %s, config=%s", name, config.Name),
				attribute.Expr.Range(),
			)
		}, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
