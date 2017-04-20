package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
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

func (d *AwsRouteNotSpecifiedTargetDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	targets := []string{
		"gateway_id",
		"egress_only_gateway_id",
		"nat_gateway_id",
		"instance_id",
		"vpc_peering_connection_id",
		"network_interface_id",
	}

	for _, t := range targets {
		if _, ok := resource.GetToken(t); ok {
			return
		}
	}

	issue := &issue.Issue{
		Type:    d.IssueType,
		Message: "route target is not specified, each route must contain either a gateway_id, egress_only_gateway_id a nat_gateway_id, an instance_id or a vpc_peering_connection_id or a network_interface_id.",
		Line:    resource.Pos.Line,
		File:    resource.Pos.Filename,
	}
	*issues = append(*issues, issue)
}
