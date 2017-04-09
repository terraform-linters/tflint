package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteNotSpecifiedTargetDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateAwsRouteNotSpecifiedTargetDetector() *AwsRouteNotSpecifiedTargetDetector {
	return &AwsRouteNotSpecifiedTargetDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_route",
		DeepCheck: false,
	}
}

func (d *AwsRouteNotSpecifiedTargetDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	if IsKeyNotFound(item, "gateway_id") &&
		IsKeyNotFound(item, "egress_only_gateway_id") &&
		IsKeyNotFound(item, "nat_gateway_id") &&
		IsKeyNotFound(item, "instance_id") &&
		IsKeyNotFound(item, "vpc_peering_connection_id") &&
		IsKeyNotFound(item, "network_interface_id") {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: "route target is not specified, each route must contain either a gateway_id, egress_only_gateway_id a nat_gateway_id, an instance_id or a vpc_peering_connection_id or a network_interface_id.",
			Line:    item.Pos().Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}

}
