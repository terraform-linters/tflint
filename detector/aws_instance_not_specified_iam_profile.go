package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceNotSpecifiedIAMProfileDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceNotSpecifiedIAMProfileDetector() *AwsInstanceNotSpecifiedIAMProfileDetector {
	nd := &AwsInstanceNotSpecifiedIAMProfileDetector{Detector: d}
	nd.Name = "aws_instance_not_specified_iam_profile"
	nd.IssueType = issue.NOTICE
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_instance_not_specified_iam_profile.md"
	return nd
}

func (d *AwsInstanceNotSpecifiedIAMProfileDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	if _, ok := resource.GetToken("iam_instance_profile"); !ok {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: "\"iam_instance_profile\" is not specified. If you want to change it, you need to recreate instance. (Only less than Terraform 0.8.8)",
			Line:    resource.Pos.Line,
			File:    resource.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
