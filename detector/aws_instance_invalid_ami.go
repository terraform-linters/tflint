package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidAMIDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceInvalidAMIDetector() *AwsInstanceInvalidAMIDetector {
	return &AwsInstanceInvalidAMIDetector{d}
}

func (d *AwsInstanceInvalidAMIDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_instance") {
		return
	}

	validAmi := map[string]bool{}
	if d.ResponseCache.DescribeImagesOutput == nil {
		resp, err := d.AwsClient.Ec2.DescribeImages(&ec2.DescribeImagesInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeImagesOutput = resp
	}
	for _, image := range d.ResponseCache.DescribeImagesOutput.Images {
		validAmi[*image.ImageId] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			amiToken, err := hclLiteralToken(item, "ami")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			ami, err := d.evalToString(amiToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validAmi[ami] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid AMI.", ami),
					Line:    amiToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
