package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidEgressOnlyGatewayDetector struct {
	*Detector
	egateways map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidEgressOnlyGatewayDetector() *AwsRouteInvalidEgressOnlyGatewayDetector {
	nd := &AwsRouteInvalidEgressOnlyGatewayDetector{
		Detector:  d,
		egateways: map[string]bool{},
	}
	nd.Name = "aws_route_invalid_egress_only_gateway"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = true
	nd.Enabled = true
	return nd
}

func (d *AwsRouteInvalidEgressOnlyGatewayDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeEgressOnlyInternetGateways()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, egateway := range resp.EgressOnlyInternetGateways {
		d.egateways[*egateway.EgressOnlyInternetGatewayId] = true
	}
}

func (d *AwsRouteInvalidEgressOnlyGatewayDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	egatewayToken, ok := resource.GetToken("egress_only_gateway_id")
	if !ok {
		return
	}
	egateway, err := d.evalToString(egatewayToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.egateways[egateway] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid egress only internet gateway ID.", egateway),
			Line:     egatewayToken.Pos.Line,
			File:     egatewayToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
