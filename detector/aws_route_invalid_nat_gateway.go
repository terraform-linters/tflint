package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidNatGatewayDetector struct {
	*Detector
	ngateways map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidNatGatewayDetector() *AwsRouteInvalidNatGatewayDetector {
	nd := &AwsRouteInvalidNatGatewayDetector{
		Detector:  d,
		ngateways: map[string]bool{},
	}
	nd.Name = "aws_route_invalid_nat_gateway"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = true
	return nd
}

func (d *AwsRouteInvalidNatGatewayDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeNatGateways()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, ngateway := range resp.NatGateways {
		d.ngateways[*ngateway.NatGatewayId] = true
	}
}

func (d *AwsRouteInvalidNatGatewayDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	ngatewayToken, ok := resource.GetToken("nat_gateway_id")
	if !ok {
		return
	}
	ngateway, err := d.evalToString(ngatewayToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.ngateways[ngateway] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid NAT gateway ID.", ngateway),
			Line:    ngatewayToken.Pos.Line,
			File:    ngatewayToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
