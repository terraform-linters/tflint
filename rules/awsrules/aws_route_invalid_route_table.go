package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteInvalidRouteTableRule checks whether route table actually exists
type AwsRouteInvalidRouteTableRule struct {
	resourceType  string
	attributeName string
	routeTables   map[string]bool
	dataPrepared  bool
}

// NewAwsRouteInvalidRouteTableRule returns new rule with default attributes
func NewAwsRouteInvalidRouteTableRule() *AwsRouteInvalidRouteTableRule {
	return &AwsRouteInvalidRouteTableRule{
		resourceType:  "aws_route",
		attributeName: "route_table_id",
		routeTables:   map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsRouteInvalidRouteTableRule) Name() string {
	return "aws_route_invalid_route_table"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteInvalidRouteTableRule) Enabled() bool {
	return true
}

// Check checks whether `route_table_id` are included in the list retrieved by `DescribeRouteTables`
func (r *AwsRouteInvalidRouteTableRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch route tables")
			resp, err := runner.AwsClient.EC2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing route tables",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, routeTable := range resp.RouteTables {
				r.routeTables[*routeTable.RouteTableId] = true
			}
			r.dataPrepared = true
		}

		var routeTable string
		err := runner.EvaluateExpr(attribute.Expr, &routeTable)

		return runner.EnsureNoError(err, func() error {
			if !r.routeTables[routeTable] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid route table ID.", routeTable),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
			return nil
		})
	})
}
