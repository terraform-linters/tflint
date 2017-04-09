package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteInvalidVpcPeeringConnectionDetector struct {
	*Detector
	IssueType             string
	Target                string
	DeepCheck             bool
	vpcPeeringConnections map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidVpcPeeringConnectionDetector() *AwsRouteInvalidVpcPeeringConnectionDetector {
	return &AwsRouteInvalidVpcPeeringConnectionDetector{
		Detector:              d,
		IssueType:             issue.ERROR,
		Target:                "aws_route",
		DeepCheck:             true,
		vpcPeeringConnections: map[string]bool{},
	}
}

func (d *AwsRouteInvalidVpcPeeringConnectionDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeVpcPeeringConnections()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, vpcPeeringConnection := range resp.VpcPeeringConnections {
		d.vpcPeeringConnections[*vpcPeeringConnection.VpcPeeringConnectionId] = true
	}
}

func (d *AwsRouteInvalidVpcPeeringConnectionDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	vpcPeeringConnectionToken, err := hclLiteralToken(item, "vpc_peering_connection_id")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	vpcPeeringConnection, err := d.evalToString(vpcPeeringConnectionToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.vpcPeeringConnections[vpcPeeringConnection] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid VPC peering connection ID.", vpcPeeringConnection),
			Line:    vpcPeeringConnectionToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
