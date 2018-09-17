package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidEgressOnlyGatewayRule checks whether egress only gateway actually exists
type AwsRouteInvalidEgressOnlyGatewayRule struct {
	resourceType  string
	attributeName string
	egateways     map[string]bool
	dataPrepared  bool
}

// NewAwsRouteInvalidEgressOnlyGatewayRule returns new rule with default attributes
func NewAwsRouteInvalidEgressOnlyGatewayRule() *AwsRouteInvalidEgressOnlyGatewayRule {
	return &AwsRouteInvalidEgressOnlyGatewayRule{
		resourceType:  "aws_route",
		attributeName: "egress_only_gateway_id",
		egateways:     map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidEgressOnlyGatewayRule) Name() string {
	return "aws_route_invalid_egress_only_gateway"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidEgressOnlyGatewayRule) Enabled() bool {
	return true
}

// Check checks whether `egress_only_gateway_id` are included in the list retrieved by `DescribeEgressOnlyInternetGateways`
func (r *AwsRouteInvalidEgressOnlyGatewayRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch egress only internet gateways")
			resp, err := runner.AwsClient.EC2.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing egress only internet gateways",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, egateway := range resp.EgressOnlyInternetGateways {
				r.egateways[*egateway.EgressOnlyInternetGatewayId] = true
			}
			r.dataPrepared = true
		}

		var egateway string
		err := runner.EvaluateExpr(attribute.Expr, &egateway)

		return runner.EnsureNoError(err, func() error {
			if !r.egateways[egateway] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid egress only internet gateway ID.", egateway),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
			return nil
		})
	})
}
