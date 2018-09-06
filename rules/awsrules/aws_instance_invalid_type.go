package awsrules

import (
	"fmt"
	"log"

	instances "github.com/cristim/ec2-instances-info"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsInstanceInvalidTypeRule checks whether "aws_instance" has invalid intance type.
type AwsInstanceInvalidTypeRule struct {
	resourceType  string
	attributeName string
	instanceTypes map[string]bool
}

// NewAwsInstanceInvalidTypeRule returns new rule with default attributes
func NewAwsInstanceInvalidTypeRule() *AwsInstanceInvalidTypeRule {
	rule := &AwsInstanceInvalidTypeRule{
		resourceType:  "aws_instance",
		attributeName: "instance_type",
		instanceTypes: map[string]bool{},
	}

	data, err := instances.Data()
	if err != nil {
		// Maybe this is bug
		panic(err)
	}

	for _, i := range *data {
		rule.instanceTypes[i.InstanceType] = true
	}

	return rule
}

// Name returns the rule name
func (r *AwsInstanceInvalidTypeRule) Name() string {
	return "aws_instance_invalid_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceInvalidTypeRule) Enabled() bool {
	return true
}

// Check checks whether "aws_instance" has invalid instance type.
func (r *AwsInstanceInvalidTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var instanceType string
		err := runner.EvaluateExpr(attribute.Expr, &instanceType)

		return runner.EnsureNoError(err, func() error {
			if !r.instanceTypes[instanceType] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				})
			}
			return nil
		})
	})
}
