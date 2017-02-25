package detector

import "github.com/wata727/tflint/issue"

type AwsInstanceNotSpecifiedIAMProfileDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateAwsInstanceNotSpecifiedIAMProfileDetector() *AwsInstanceNotSpecifiedIAMProfileDetector {
	return &AwsInstanceNotSpecifiedIAMProfileDetector{
		Detector:  d,
		IssueType: issue.NOTICE,
		Target:    "aws_instance",
		DeepCheck: false,
	}
}

func (d *AwsInstanceNotSpecifiedIAMProfileDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			if IsKeyNotFound(item, "iam_instance_profile") {
				issue := &issue.Issue{
					Type:    "NOTICE",
					Message: "\"iam_instance_profile\" is not specified. If you want to change it, you need to recreate it",
					Line:    item.Pos().Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
