package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidVPCPeeringConnectionRule checks whether VPC peering connection actually exists
type AwsRouteInvalidVPCPeeringConnectionRule struct {
	resourceType          string
	attributeName         string
	vpcPeeringConnections map[string]bool
	dataPrepared          bool
}

// NewAwsRouteInvalidVPCPeeringConnectionRule returns new rule with default attributes
func NewAwsRouteInvalidVPCPeeringConnectionRule() *AwsRouteInvalidVPCPeeringConnectionRule {
	return &AwsRouteInvalidVPCPeeringConnectionRule{
		resourceType:          "aws_route",
		attributeName:         "vpc_peering_connection_id",
		vpcPeeringConnections: map[string]bool{},
		dataPrepared:          false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidVPCPeeringConnectionRule) Name() string {
	return "aws_route_invalid_vpc_peering_connection"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidVPCPeeringConnectionRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsRouteInvalidVPCPeeringConnectionRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsRouteInvalidVPCPeeringConnectionRule) Link() string {
	return ""
}

// Check checks whether `vpc_peering_connection_id` are included in the list retrieved by `DescribeVpcPeeringConnections`
func (r *AwsRouteInvalidVPCPeeringConnectionRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch VPC peering connections")
			resp, err := runner.AwsClient.EC2.DescribeVpcPeeringConnections(&ec2.DescribeVpcPeeringConnectionsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing VPC peering connections",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, vpcPeeringConnection := range resp.VpcPeeringConnections {
				r.vpcPeeringConnections[*vpcPeeringConnection.VpcPeeringConnectionId] = true
			}
			r.dataPrepared = true
		}

		var vpcPeeringConnection string
		err := runner.EvaluateExpr(attribute.Expr, &vpcPeeringConnection)

		return runner.EnsureNoError(err, func() error {
			if !r.vpcPeeringConnections[vpcPeeringConnection] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid VPC peering connection ID.", vpcPeeringConnection),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
