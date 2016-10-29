package detector

import (
	"reflect"

	"github.com/hashicorp/hcl/hcl/ast"
	eval "github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

type Detector struct {
	ListMap    map[string]*ast.ObjectList
	EvalConfig *eval.Evaluator
}

var detectors = []string{
	"DetectAwsInstanceInvalidType",
	"DetectAwsInstancePreviousType",
	"DetectAwsInstanceNotSpecifiedIamProfile",
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

	for _, detectorMethod := range detectors {
		method := reflect.ValueOf(d).MethodByName(detectorMethod)
		method.Call([]reflect.Value{reflect.ValueOf(&issues)})

		for _, m := range d.EvalConfig.ModuleConfig {
			moduleDetector := &Detector{
				ListMap: m.ListMap,
				EvalConfig: &eval.Evaluator{
					Config: m.Config,
				},
			}
			method := reflect.ValueOf(moduleDetector).MethodByName(detectorMethod)
			method.Call([]reflect.Value{reflect.ValueOf(&issues)})
		}
	}

	return issues
}
