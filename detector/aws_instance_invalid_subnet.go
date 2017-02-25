package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidSubnetDetector struct {
	*Detector
	subnets map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidSubnetDetector() *AwsInstanceInvalidSubnetDetector {
	return &AwsInstanceInvalidSubnetDetector{
		Detector: d,
		subnets:  map[string]bool{},
	}
}

func (d *AwsInstanceInvalidSubnetDetector) PreProcess() {
	if d.isSkippable("resource", "aws_instance") {
		return
	}

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

func (d *AwsInstanceInvalidSubnetDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			subnetToken, err := hclLiteralToken(item, "subnet_id")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			subnet, err := d.evalToString(subnetToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !d.subnets[subnet] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid subnet ID.", subnet),
					Line:    subnetToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
