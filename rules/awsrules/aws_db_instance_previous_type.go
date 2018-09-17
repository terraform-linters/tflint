package awsrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstancePreviousTypeRule checks whether the resource uses previous generation instance type
type AwsDBInstancePreviousTypeRule struct {
	resourceType          string
	attributeName         string
	previousInstanceTypes map[string]bool
}

// NewAwsDBInstancePreviousTypeRule returns new rule with default attributes
func NewAwsDBInstancePreviousTypeRule() *AwsDBInstancePreviousTypeRule {
	return &AwsDBInstancePreviousTypeRule{
		resourceType:  "aws_db_instance",
		attributeName: "instance_class",
		previousInstanceTypes: map[string]bool{
			"db.t1.micro":    true,
			"db.m1.small":    true,
			"db.m1.medium":   true,
			"db.m1.large":    true,
			"db.m1.xlarge":   true,
			"db.m2.xlarge":   true,
			"db.m2.2xlarge":  true,
			"db.m2.4xlarge":  true,
			"db.cr1.8xlarge": true,
		},
	}
}

// Name returns the rule name
func (r *AwsDBInstancePreviousTypeRule) Name() string {
	return "aws_db_instance_previous_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstancePreviousTypeRule) Enabled() bool {
	return true
}

// Check checks whether the resource's `instance_class` is included in the list of previous generation instance type
func (r *AwsDBInstancePreviousTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var instanceType string
		err := runner.EvaluateExpr(attribute.Expr, &instanceType)

		return runner.EnsureNoError(err, func() error {
			if r.previousInstanceTypes[instanceType] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.WARNING,
					Message:  fmt.Sprintf("\"%s\" is previous generation instance type.", instanceType),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_previous_type.md",
				})
			}
			return nil
		})
	})
}
