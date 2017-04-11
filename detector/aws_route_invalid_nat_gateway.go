package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteInvalidNatGatewayDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	ngateways map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidNatGatewayDetector() *AwsRouteInvalidNatGatewayDetector {
	return &AwsRouteInvalidNatGatewayDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_route",
		DeepCheck: true,
		ngateways: map[string]bool{},
	}
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

func (d *AwsRouteInvalidNatGatewayDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	ngatewayToken, err := hclLiteralToken(item, "nat_gateway_id")
	if err != nil {
		d.Logger.Error(err)
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
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
