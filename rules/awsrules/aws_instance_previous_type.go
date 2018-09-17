package awsrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsInstancePreviousTypeRule checks whether the resource uses previous generation instance type
type AwsInstancePreviousTypeRule struct {
	resourceType          string
	attributeName         string
	previousInstanceTypes map[string]bool
}

// NewAwsInstancePreviousTypeRule returns new rule with default attributes
func NewAwsInstancePreviousTypeRule() *AwsInstancePreviousTypeRule {
	return &AwsInstancePreviousTypeRule{
		resourceType:  "aws_instance",
		attributeName: "instance_type",
		previousInstanceTypes: map[string]bool{
			"t1.micro":    true,
			"m1.small":    true,
			"m1.medium":   true,
			"m1.large":    true,
			"m1.xlarge":   true,
			"c1.medium":   true,
			"c1.xlarge":   true,
			"cc2.8xlarge": true,
			"cg1.4xlarge": true,
			"m2.xlarge":   true,
			"m2.2xlarge":  true,
			"m2.4xlarge":  true,
			"cr1.8xlarge": true,
			"hi1.4xlarge": true,
			"hs1.8xlarge": true,
		},
	}
}

// Name returns the rule name
func (r *AwsInstancePreviousTypeRule) Name() string {
	return "aws_instance_previous_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstancePreviousTypeRule) Enabled() bool {
	return true
}

// Check checks whether the resource's `instance_type` is included in the list of previous generation instance type
func (r *AwsInstancePreviousTypeRule) Check(runner *tflint.Runner) error {
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
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_previous_type.md",
				})
			}
			return nil
		})
	})
}
