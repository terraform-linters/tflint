package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteSpecifiedMultipleTargetsDetector struct {
	*Detector
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
}

func (d *Detector) CreateAwsRouteSpecifiedMultipleTargetsDetector() *AwsRouteSpecifiedMultipleTargetsDetector {
	return &AwsRouteSpecifiedMultipleTargetsDetector{
		Detector:   d,
		IssueType:  issue.ERROR,
		TargetType: "resource",
		Target:     "aws_route",
		DeepCheck:  false,
	}
}

func (d *AwsRouteSpecifiedMultipleTargetsDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	targets := []string{"gateway_id", "egress_only_gateway_id", "nat_gateway_id", "instance_id", "vpc_peering_connection_id", "network_interface_id"}

	var targetCount int = 0
	for _, target := range targets {
		if _, ok := resource.GetToken(target); ok {
			targetCount++
			if targetCount > 1 {
				issue := &issue.Issue{
					Type:    d.IssueType,
					Message: "more than 1 target specified, only 1 routing target can be specified.",
					Line:    resource.Pos.Line,
					File:    resource.Pos.Filename,
				}
				*issues = append(*issues, issue)
				return
			}
		}
	}
}
