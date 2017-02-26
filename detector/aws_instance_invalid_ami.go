package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
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

func (d *AwsInstanceInvalidAMIDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	amiToken, err := hclLiteralToken(item, "ami")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	ami, err := d.evalToString(amiToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.amis[ami] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid AMI.", ami),
			Line:    amiToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
