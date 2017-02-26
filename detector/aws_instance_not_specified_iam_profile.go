package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

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

func (d *AwsInstanceNotSpecifiedIAMProfileDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	if IsKeyNotFound(item, "iam_instance_profile") {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: "\"iam_instance_profile\" is not specified. If you want to change it, you need to recreate it",
			Line:    item.Pos().Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
