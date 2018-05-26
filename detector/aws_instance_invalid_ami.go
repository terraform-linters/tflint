package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceInvalidAMIDetector struct {
	*Detector
	amis map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidAMIDetector() *AwsInstanceInvalidAMIDetector {
	nd := &AwsInstanceInvalidAMIDetector{
		Detector: d,
		amis:     map[string]bool{},
	}
	nd.Name = "aws_instance_invalid_ami"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = true
	nd.Enabled = true
	return nd
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

func (d *AwsInstanceInvalidAMIDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	amiToken, ok := resource.GetToken("ami")
	if !ok {
		return
	}
	ami, err := d.evalToString(amiToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.amis[ami] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid AMI.", ami),
			Line:     amiToken.Pos.Line,
			File:     amiToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
