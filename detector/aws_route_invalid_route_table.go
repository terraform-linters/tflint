package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteInvalidRouteTableDetector struct {
	*Detector
	IssueType   string
	Target      string
	DeepCheck   bool
	routeTables map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidRouteTableDetector() *AwsRouteInvalidRouteTableDetector {
	return &AwsRouteInvalidRouteTableDetector{
		Detector:    d,
		IssueType:   issue.ERROR,
		Target:      "aws_route",
		DeepCheck:   true,
		routeTables: map[string]bool{},
	}
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

func (d *AwsRouteInvalidRouteTableDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	routeTableToken, err := hclLiteralToken(item, "route_table_id")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	routeTable, err := d.evalToString(routeTableToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.routeTables[routeTable] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid route table ID.", routeTable),
			Line:    routeTableToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
