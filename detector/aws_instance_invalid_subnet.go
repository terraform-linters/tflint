package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidSubnetDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceInvalidSubnetDetector() *AwsInstanceInvalidSubnetDetector {
	return &AwsInstanceInvalidSubnetDetector{d}
}

func (d *AwsInstanceInvalidSubnetDetector) Detect(issues *[]*issue.Issue) {
	if !d.Config.DeepCheck {
		d.Logger.Info("skip this rule.")
		return
	}

	validSubnets := map[string]bool{}
	if d.ResponseCache.DescribeSubnetsOutput == nil {
		resp, err := d.AwsClient.Ec2.DescribeSubnets(&ec2.DescribeSubnetsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeSubnetsOutput = resp
	}
	for _, subnet := range d.ResponseCache.DescribeSubnetsOutput.Subnets {
		validSubnets[*subnet.SubnetId] = true
	}

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

			if !validSubnets[subnet] {
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
