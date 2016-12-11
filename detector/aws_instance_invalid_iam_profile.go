package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidIAMProfileDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceInvalidIAMProfileDetector() *AwsInstanceInvalidIAMProfileDetector {
	return &AwsInstanceInvalidIAMProfileDetector{d}
}

func (d *AwsInstanceInvalidIAMProfileDetector) Detect(issues *[]*issue.Issue) {
	if !d.Config.DeepCheck {
		d.Logger.Info("skip this rule.")
		return
	}

	validIamProfiles := map[string]bool{}
	resp, err := d.AwsClient.Iam.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
	}
	for _, iamProfile := range resp.InstanceProfiles {
		validIamProfiles[*iamProfile.InstanceProfileName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			iamProfileToken, err := hclLiteralToken(item, "iam_instance_profile")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			iamProfile, err := d.evalToString(iamProfileToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validIamProfiles[iamProfile] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid IAM profile name.", iamProfile),
					Line:    iamProfileToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
