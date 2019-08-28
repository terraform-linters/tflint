package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsELBInvalidInstanceRule checks whether instances actually exists
type AwsELBInvalidInstanceRule struct {
	resourceType  string
	attributeName string
	instances     map[string]bool
	dataPrepared  bool
}

// NewAwsELBInvalidInstanceRule returns new rule with default attributes
func NewAwsELBInvalidInstanceRule() *AwsELBInvalidInstanceRule {
	return &AwsELBInvalidInstanceRule{
		resourceType:  "aws_elb",
		attributeName: "instances",
		instances:     map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsELBInvalidInstanceRule) Name() string {
	return "aws_elb_invalid_instance"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsELBInvalidInstanceRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsELBInvalidInstanceRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsELBInvalidInstanceRule) Link() string {
	return ""
}

// Check checks whether `instances` are included in the list retrieved by `DescribeInstances`
func (r *AwsELBInvalidInstanceRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch instances")
			resp, err := runner.AwsClient.EC2.DescribeInstances(&ec2.DescribeInstancesInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing instances",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, reservation := range resp.Reservations {
				for _, instance := range reservation.Instances {
					r.instances[*instance.InstanceId] = true
				}
			}
			r.dataPrepared = true
		}

		return runner.EachStringSliceExprs(attribute.Expr, func(instance string, expr hcl.Expression) {
			if !r.instances[instance] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid instance.", instance),
					expr.Range(),
				)
			}
		})
	})
}
