package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidGatewayDetector struct {
	*Detector
	gateways map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidGatewayDetector() *AwsRouteInvalidGatewayDetector {
	nd := &AwsRouteInvalidGatewayDetector{
		Detector: d,
		gateways: map[string]bool{},
	}
	nd.Name = "aws_route_invalid_gateway"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = true
	return nd
}

func (d *AwsRouteInvalidGatewayDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeInternetGateways()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, gateway := range resp.InternetGateways {
		d.gateways[*gateway.InternetGatewayId] = true
	}
}

func (d *AwsRouteInvalidGatewayDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	gatewayToken, ok := resource.GetToken("gateway_id")
	if !ok {
		return
	}
	gateway, err := d.evalToString(gatewayToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.gateways[gateway] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid internet gateway ID.", gateway),
			Line:    gatewayToken.Pos.Line,
			File:    gatewayToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
