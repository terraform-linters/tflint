package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteSpecifiedMultipleTargetsDetector struct {
	*Detector
}

func (d *Detector) CreateAwsRouteSpecifiedMultipleTargetsDetector() *AwsRouteSpecifiedMultipleTargetsDetector {
	nd := &AwsRouteSpecifiedMultipleTargetsDetector{Detector: d}
	nd.Name = "aws_route_specified_multiple_targets"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_route_specified_multiple_targets.md"
	return nd
}

func (d *AwsRouteSpecifiedMultipleTargetsDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	targets := []string{"gateway_id", "egress_only_gateway_id", "nat_gateway_id", "instance_id", "vpc_peering_connection_id", "network_interface_id"}

	var targetCount int = 0
	for _, target := range targets {
		if _, ok := resource.GetToken(target); ok {
			targetCount++
			if targetCount > 1 {
				issue := &issue.Issue{
					Detector: d.Name,
					Type:     d.IssueType,
					Message:  "more than 1 target specified, only 1 routing target can be specified.",
					Line:     resource.Pos.Line,
					File:     resource.Pos.Filename,
					Link:     d.Link,
				}
				*issues = append(*issues, issue)
				return
			}
		}
	}
}
