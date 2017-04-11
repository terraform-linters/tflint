package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteSpecifiedMultipleTargetsDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateAwsRouteSpecifiedMultipleTargetsDetector() *AwsRouteSpecifiedMultipleTargetsDetector {
	return &AwsRouteSpecifiedMultipleTargetsDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_route",
		DeepCheck: false,
	}
}

func (d *AwsRouteSpecifiedMultipleTargetsDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	targets := []string{"gateway_id", "egress_only_gateway_id", "nat_gateway_id", "instance_id", "vpc_peering_connection_id", "network_interface_id"}

	var targetCount int = 0
	for _, target := range targets {
		if _, err := hclLiteralToken(item, target); err == nil {
			targetCount++
			if targetCount > 1 {
				issue := &issue.Issue{
					Type:    d.IssueType,
					Message: "more than 1 target specified, only 1 routing target can be specified.",
					Line:    item.Pos().Line,
					File:    file,
				}
				*issues = append(*issues, issue)
				return
			}
		}
	}
}
