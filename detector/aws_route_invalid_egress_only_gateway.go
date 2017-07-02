package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidEgressOnlyGatewayDetector struct {
	*Detector
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
	egateways  map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidEgressOnlyGatewayDetector() *AwsRouteInvalidEgressOnlyGatewayDetector {
	return &AwsRouteInvalidEgressOnlyGatewayDetector{
		Detector:   d,
		IssueType:  issue.ERROR,
		TargetType: "resource",
		Target:     "aws_route",
		DeepCheck:  true,
		egateways:  map[string]bool{},
	}
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
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid egress only internet gateway ID.", egateway),
			Line:    egatewayToken.Pos.Line,
			File:    egatewayToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
