package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsInstanceInvalidSubnetRule checks whether subnet actually exists
type AwsInstanceInvalidSubnetRule struct {
	resourceType  string
	attributeName string
	subnets       map[string]bool
	dataPrepared  bool
}

// NewAwsInstanceInvalidSubnetRule returns new rule with default attributes
func NewAwsInstanceInvalidSubnetRule() *AwsInstanceInvalidSubnetRule {
	return &AwsInstanceInvalidSubnetRule{
		resourceType:  "aws_instance",
		attributeName: "subnet_id",
		subnets:       map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsInstanceInvalidSubnetRule) Name() string {
	return "aws_instance_invalid_subnet"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceInvalidSubnetRule) Enabled() bool {
	return true
}

// Check checks whether `subnet_id` are included in the list retrieved by `DescribeSubnets`
func (r *AwsInstanceInvalidSubnetRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch subnets")
			resp, err := runner.AwsClient.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing subnets",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, subnet := range resp.Subnets {
				r.subnets[*subnet.SubnetId] = true
			}
			r.dataPrepared = true
		}

		var subnet string
		err := runner.EvaluateExpr(attribute.Expr, &subnet)

		return runner.EnsureNoError(err, func() error {
			if !r.subnets[subnet] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid subnet ID.", subnet),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
			return nil
		})
	})
}
