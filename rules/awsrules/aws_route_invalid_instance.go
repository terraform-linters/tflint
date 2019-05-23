package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidInstanceRule checks whether instance actually exists
type AwsRouteInvalidInstanceRule struct {
	resourceType  string
	attributeName string
	instances     map[string]bool
	dataPrepared  bool
}

// NewAwsRouteInvalidInstanceRule returns new rule with default attributes
func NewAwsRouteInvalidInstanceRule() *AwsRouteInvalidInstanceRule {
	return &AwsRouteInvalidInstanceRule{
		resourceType:  "aws_route",
		attributeName: "instance_id",
		instances:     map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidInstanceRule) Name() string {
	return "aws_route_invalid_instance"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidInstanceRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsRouteInvalidInstanceRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsRouteInvalidInstanceRule) Link() string {
	return ""
}

// Check checks whether `instance_id` are included in the list retrieved by `DescribeInstances`
func (r *AwsRouteInvalidInstanceRule) Check(runner *tflint.Runner) error {
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

		var instance string
		err := runner.EvaluateExpr(attribute.Expr, &instance)

		return runner.EnsureNoError(err, func() error {
			if !r.instances[instance] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid instance ID.", instance),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
