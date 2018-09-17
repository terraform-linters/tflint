package awsrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstanceDefaultParameterGroupRule checks whether the db instance use default parameter group
type AwsDBInstanceDefaultParameterGroupRule struct {
	resourceType  string
	attributeName string
}

// NewAwsDBInstanceDefaultParameterGroupRule returns new rule with default attributes
func NewAwsDBInstanceDefaultParameterGroupRule() *AwsDBInstanceDefaultParameterGroupRule {
	return &AwsDBInstanceDefaultParameterGroupRule{
		resourceType:  "aws_db_instance",
		attributeName: "parameter_group_name",
	}
}

// Name returns the rule name
func (r *AwsDBInstanceDefaultParameterGroupRule) Name() string {
	return "aws_db_instance_default_parameter_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceDefaultParameterGroupRule) Enabled() bool {
	return true
}

var defaultDBParameterGroupRegexp = regexp.MustCompile("^default")

// Check checks the parameter group name starts with `default`
func (r *AwsDBInstanceDefaultParameterGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var name string
		err := runner.EvaluateExpr(attribute.Expr, &name)

		return runner.EnsureNoError(err, func() error {
			if defaultDBParameterGroupRegexp.Match([]byte(name)) {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.NOTICE,
					Message:  fmt.Sprintf("\"%s\" is default parameter group. You cannot edit it.", name),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_default_parameter_group.md",
				})
			}
			return nil
		})
	})
}
