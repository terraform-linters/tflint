package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteInvalidGatewayDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	gateways  map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidGatewayDetector() *AwsRouteInvalidGatewayDetector {
	return &AwsRouteInvalidGatewayDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_route",
		DeepCheck: true,
		gateways:  map[string]bool{},
	}
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

func (d *AwsRouteInvalidGatewayDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	gatewayToken, err := hclLiteralToken(item, "gateway_id")
	if err != nil {
		d.Logger.Error(err)
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
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
