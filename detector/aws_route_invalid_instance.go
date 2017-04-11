package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsRouteInvalidInstanceDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	instances map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidInstanceDetector() *AwsRouteInvalidInstanceDetector {
	return &AwsRouteInvalidInstanceDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_route",
		DeepCheck: true,
		instances: map[string]bool{},
	}
}

func (d *AwsRouteInvalidInstanceDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeInstances()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			d.instances[*instance.InstanceId] = true
		}
	}
}

func (d *AwsRouteInvalidInstanceDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	instanceToken, err := hclLiteralToken(item, "instance_id")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	instance, err := d.evalToString(instanceToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.instances[instance] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid instance ID.", instance),
			Line:    instanceToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
