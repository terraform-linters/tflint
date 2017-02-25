package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidAMIDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	amis      map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidAMIDetector() *AwsInstanceInvalidAMIDetector {
	return &AwsInstanceInvalidAMIDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_instance",
		DeepCheck: true,
		amis:      map[string]bool{},
	}
}

func (d *AwsInstanceInvalidAMIDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeImages()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, image := range resp.Images {
		d.amis[*image.ImageId] = true
	}
}

func (d *AwsInstanceInvalidAMIDetector) Detect(issues *[]*issue.Issue) {
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

			if !d.amis[ami] {
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
