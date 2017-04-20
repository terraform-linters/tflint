package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
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

func (d *AwsRouteInvalidVpcPeeringConnectionDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	vpcPeeringConnectionToken, ok := resource.GetToken("vpc_peering_connection_id")
	if !ok {
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
			File:    vpcPeeringConnectionToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
