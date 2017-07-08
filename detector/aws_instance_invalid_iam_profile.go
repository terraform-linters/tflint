package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceInvalidIAMProfileDetector struct {
	*Detector
	profiles map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidIAMProfileDetector() *AwsInstanceInvalidIAMProfileDetector {
	nd := &AwsInstanceInvalidIAMProfileDetector{
		Detector: d,
		profiles: map[string]bool{},
	}
	nd.Name = "aws_instance_invalid_iam_profile"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = true
	return nd
}

func (d *AwsInstanceInvalidIAMProfileDetector) PreProcess() {
	resp, err := d.AwsClient.ListInstanceProfiles()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, iamProfile := range resp.InstanceProfiles {
		d.profiles[*iamProfile.InstanceProfileName] = true
	}
}

func (d *AwsInstanceInvalidIAMProfileDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	iamProfileToken, ok := resource.GetToken("iam_instance_profile")
	if !ok {
		return
	}
	iamProfile, err := d.evalToString(iamProfileToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.profiles[iamProfile] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid IAM profile name.", iamProfile),
			Line:    iamProfileToken.Pos.Line,
			File:    iamProfileToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
