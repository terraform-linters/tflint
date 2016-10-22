package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/detector/aws"
	eval "github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

type Detector struct {
	List       *ast.ObjectList
	File       string
	EvalConfig *eval.Evaluator
}

func Detect(list *ast.ObjectList, file string) []*issue.Issue {
	detector := &Detector{
		List:       list,
		File:       file,
		EvalConfig: eval.NewEvaluator(list),
	}
	return detector.Detect()
}

func (d *Detector) Detect() []*issue.Issue {
	var issues = []*issue.Issue{}
	awsDetector := &aws.AwsDetector{
		List:       d.List,
		File:       d.File,
		EvalConfig: d.EvalConfig,
	}

	issues = append(issues, awsDetector.DetectAwsInstanceInvalidType()...)
	issues = append(issues, awsDetector.DetectAwsInstancePreviousType()...)
	issues = append(issues, awsDetector.DetectAwsInstanceNotSpecifiedIamProfile()...)

	return issues
}
