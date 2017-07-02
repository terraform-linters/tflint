package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceNotSpecifiedIAMProfileDetector struct {
	*Detector
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
}

func (d *Detector) CreateAwsInstanceNotSpecifiedIAMProfileDetector() *AwsInstanceNotSpecifiedIAMProfileDetector {
	return &AwsInstanceNotSpecifiedIAMProfileDetector{
		Detector:   d,
		IssueType:  issue.NOTICE,
		TargetType: "resource",
		Target:     "aws_instance",
		DeepCheck:  false,
	}
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
