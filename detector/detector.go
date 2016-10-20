package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/detector/aws"
	"github.com/wata727/tflint/issue"
)

type Detector struct {
	List *ast.ObjectList
	File string
}

func Detect(list *ast.ObjectList, file string) []*issue.Issue {
	detector := &Detector{
		List: list,
		File: file,
	}
	return detector.Detect()
}

func (d *Detector) Detect() []*issue.Issue {
	var issues = []*issue.Issue{}

	issues = append(issues, aws.DetectAwsInstanceInvalidType(d.List, d.File)...)
	issues = append(issues, aws.DetectAwsInstancePreviousType(d.List, d.File)...)
	issues = append(issues, aws.DetectAwsInstanceNotSpecifiedIamProfile(d.List, d.File)...)

	return issues
}
