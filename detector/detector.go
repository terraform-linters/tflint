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

func Detect(listmap map[string]*ast.ObjectList) ([]*issue.Issue, error) {
	evalConfig, err := eval.NewEvaluator(listmap)
	if err != nil {
		return nil, err
	}

	detector := &Detector{
		ListMap:    listmap,
		EvalConfig: evalConfig,
	}
	return detector.Detect(), nil
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

	for _, m := range d.EvalConfig.ModuleConfig {
		awsModuleDetector := &aws.AwsDetector{
			ListMap: m.ListMap,
			EvalConfig: &eval.Evaluator{
				Config: m.Config,
			},
		}

		issues = append(issues, awsModuleDetector.DetectAwsInstanceInvalidType()...)
		issues = append(issues, awsModuleDetector.DetectAwsInstancePreviousType()...)
		issues = append(issues, awsModuleDetector.DetectAwsInstanceNotSpecifiedIamProfile()...)
	}

	return issues
}
