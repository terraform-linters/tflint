package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidRouteTableDetector struct {
	*Detector
	routeTables map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidRouteTableDetector() *AwsRouteInvalidRouteTableDetector {
	nd := &AwsRouteInvalidRouteTableDetector{
		Detector:    d,
		routeTables: map[string]bool{},
	}
	nd.Name = "aws_route_invalid_route_table"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = true
	return nd
}

func (d *AwsRouteInvalidRouteTableDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeRouteTables()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, routeTable := range resp.RouteTables {
		d.routeTables[*routeTable.RouteTableId] = true
	}
}

func (d *AwsRouteInvalidRouteTableDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	routeTableToken, ok := resource.GetToken("route_table_id")
	if !ok {
		return
	}
	routeTable, err := d.evalToString(routeTableToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.routeTables[routeTable] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid route table ID.", routeTable),
			Line:     routeTableToken.Pos.Line,
			File:     routeTableToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
