package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidNetworkInterfaceRule checks whether network interface actually exists
type AwsRouteInvalidNetworkInterfaceRule struct {
	resourceType      string
	attributeName     string
	networkInterfaces map[string]bool
	dataPrepared      bool
}

// NewAwsRouteInvalidNetworkInterfaceRule returns new rule with default attributes
func NewAwsRouteInvalidNetworkInterfaceRule() *AwsRouteInvalidNetworkInterfaceRule {
	return &AwsRouteInvalidNetworkInterfaceRule{
		resourceType:      "aws_route",
		attributeName:     "network_interface_id",
		networkInterfaces: map[string]bool{},
		dataPrepared:      false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidNetworkInterfaceRule) Name() string {
	return "aws_route_invalid_network_interface"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidNetworkInterfaceRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsRouteInvalidNetworkInterfaceRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsRouteInvalidNetworkInterfaceRule) Link() string {
	return ""
}

// Check checks whether `network_interface_id` are included in the list retrieved by `DescribeNetworkInterfaces`
func (r *AwsRouteInvalidNetworkInterfaceRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch network interfaces")
			resp, err := runner.AwsClient.EC2.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing network interfaces",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, networkInterface := range resp.NetworkInterfaces {
				r.networkInterfaces[*networkInterface.NetworkInterfaceId] = true
			}
			r.dataPrepared = true
		}

		var networkInterface string
		err := runner.EvaluateExpr(attribute.Expr, &networkInterface)

		return runner.EnsureNoError(err, func() error {
			if !r.networkInterfaces[networkInterface] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid network interface ID.", networkInterface),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
