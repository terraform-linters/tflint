package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/detector/aws"
	eval "github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

type Detector struct {
	ListMap    map[string]*ast.ObjectList
	EvalConfig *eval.Evaluator
}

func Detect(listmap map[string]*ast.ObjectList) []*issue.Issue {
	detector := &Detector{
		ListMap:    listmap,
		EvalConfig: eval.NewEvaluator(listmap),
	}
	return detector.Detect()
}

func (d *Detector) Detect() []*issue.Issue {
	var issues = []*issue.Issue{}
	awsDetector := &aws.AwsDetector{
		ListMap:    d.ListMap,
		EvalConfig: d.EvalConfig,
	}

	issues = append(issues, awsDetector.DetectAwsInstanceInvalidType()...)
	issues = append(issues, awsDetector.DetectAwsInstancePreviousType()...)
	issues = append(issues, awsDetector.DetectAwsInstanceNotSpecifiedIamProfile()...)

	return issues
}
