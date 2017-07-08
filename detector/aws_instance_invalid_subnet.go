package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceInvalidSubnetDetector struct {
	*Detector
	subnets map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidSubnetDetector() *AwsInstanceInvalidSubnetDetector {
	nd := &AwsInstanceInvalidSubnetDetector{
		Detector: d,
		subnets:  map[string]bool{},
	}
	nd.Name = "aws_instance_invalid_subnet"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = true
	return nd
}

func (d *AwsInstanceInvalidSubnetDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeSubnets()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, subnet := range resp.Subnets {
		d.subnets[*subnet.SubnetId] = true
	}
}

func (d *AwsInstanceInvalidSubnetDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	subnetToken, ok := resource.GetToken("subnet_id")
	if !ok {
		return
	}
	subnet, err := d.evalToString(subnetToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.subnets[subnet] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid subnet ID.", subnet),
			Line:     subnetToken.Pos.Line,
			File:     subnetToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
