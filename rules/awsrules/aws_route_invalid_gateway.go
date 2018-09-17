package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidGatewayRule checks whether internet gateway actually exists
type AwsRouteInvalidGatewayRule struct {
	resourceType  string
	attributeName string
	gateways      map[string]bool
	dataPrepared  bool
}

// NewAwsRouteInvalidGatewayRule returns new rule with default attributes
func NewAwsRouteInvalidGatewayRule() *AwsRouteInvalidGatewayRule {
	return &AwsRouteInvalidGatewayRule{
		resourceType:  "aws_route",
		attributeName: "gateway_id",
		gateways:      map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidGatewayRule) Name() string {
	return "aws_route_invalid_gateway"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidGatewayRule) Enabled() bool {
	return true
}

// Check checks whether `gateway_id` are included in the list retrieved by `DescribeInternetGateways`
func (r *AwsRouteInvalidGatewayRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch internet gateways")
			resp, err := runner.AwsClient.EC2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing internet gateways",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, gateway := range resp.InternetGateways {
				r.gateways[*gateway.InternetGatewayId] = true
			}
			r.dataPrepared = true
		}

		var gateway string
		err := runner.EvaluateExpr(attribute.Expr, &gateway)

		return runner.EnsureNoError(err, func() error {
			if !r.gateways[gateway] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid internet gateway ID.", gateway),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
			return nil
		})
	})
}
