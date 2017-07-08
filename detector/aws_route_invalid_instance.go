package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidInstanceDetector struct {
	*Detector
	instances map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidInstanceDetector() *AwsRouteInvalidInstanceDetector {
	nd := &AwsRouteInvalidInstanceDetector{
		Detector:  d,
		instances: map[string]bool{},
	}
	nd.Name = "aws_route_invalid_instance"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_route"
	nd.DeepCheck = true
	return nd
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

func (d *AwsRouteInvalidInstanceDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	instanceToken, ok := resource.GetToken("instance_id")
	if !ok {
		return
	}
	instance, err := d.evalToString(instanceToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.instances[instance] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid instance ID.", instance),
			Line:     instanceToken.Pos.Line,
			File:     instanceToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
