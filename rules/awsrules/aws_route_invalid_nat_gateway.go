package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidNatGatewayRule checks whether NAT gateway actually exists
type AwsRouteInvalidNatGatewayRule struct {
	resourceType  string
	attributeName string
	ngateways     map[string]bool
	dataPrepared  bool
}

// NewAwsRouteInvalidNatGatewayRule returns new rule with default attributes
func NewAwsRouteInvalidNatGatewayRule() *AwsRouteInvalidNatGatewayRule {
	return &AwsRouteInvalidNatGatewayRule{
		resourceType:  "aws_route",
		attributeName: "nat_gateway_id",
		ngateways:     map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidNatGatewayRule) Name() string {
	return "aws_route_invalid_nat_gateway"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidNatGatewayRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsRouteInvalidNatGatewayRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsRouteInvalidNatGatewayRule) Link() string {
	return ""
}

// Check checks whether `nat_gateway_id` are included in the list retrieved by `DescribeNatGateways`
func (r *AwsRouteInvalidNatGatewayRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch NAT gateways")
			resp, err := runner.AwsClient.EC2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing NAT gateways",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, ngateway := range resp.NatGateways {
				r.ngateways[*ngateway.NatGatewayId] = true
			}
			r.dataPrepared = true
		}

		var ngateway string
		err := runner.EvaluateExpr(attribute.Expr, &ngateway)

		return runner.EnsureNoError(err, func() error {
			if !r.ngateways[ngateway] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid NAT gateway ID.", ngateway),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
