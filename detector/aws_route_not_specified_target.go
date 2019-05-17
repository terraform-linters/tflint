package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteNotSpecifiedTargetDetector struct {
	*Detector
}

func (d *Detector) CreateAwsRouteNotSpecifiedTargetDetector() *AwsRouteNotSpecifiedTargetDetector {
	nd := &AwsRouteNotSpecifiedTargetDetector{Detector: d}
	nd.Name = "aws_route_not_specified_target"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_route_not_specified_target.md"
	nd.Enabled = true
	return nd
}

func (d *AwsRouteNotSpecifiedTargetDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	targets := []string{
		"gateway_id",
		"egress_only_gateway_id",
		"nat_gateway_id",
		"instance_id",
		"vpc_peering_connection_id",
		"network_interface_id",
		"transit_gateway_id",
	}

	for _, t := range targets {
		if _, ok := resource.GetToken(t); ok {
			return
		}
	}

	issue := &issue.Issue{
		Detector: d.Name,
		Type:     d.IssueType,
		Message:  "The routing target is not specified, each aws_route must contain either egress_only_gateway_id, gateway_id, instance_id, nat_gateway_id, network_interface_id, transit_gateway_id, or vpc_peering_connection_id.",
		Line:     resource.Pos.Line,
		File:     resource.Pos.Filename,
		Link:     d.Link,
	}
	*issues = append(*issues, issue)
}
