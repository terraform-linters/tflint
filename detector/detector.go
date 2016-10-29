package detector

import (
	"errors"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
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
	return detector.detect(), nil
}

func hclLiteralToken(item *ast.ObjectItem, k string) (token.Token, error) {
	items := item.Val.(*ast.ObjectType).List.Filter(k).Items
	if len(items) == 0 {
		return token.Token{}, errors.New("key not found")
	}

	v := items[0].Val
	switch v.(type) {
	case *ast.LiteralType:
		return v.(*ast.LiteralType).Token, nil
	default:
		return token.Token{}, errors.New("value is not literal")
	}
}

func (d *Detector) detect() []*issue.Issue {
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

func (d *Detector) evalToString(v string) (string, error) {
	ev, err := d.EvalConfig.Eval(strings.Trim(v, "\""))

	if err != nil {
		return "", err
	} else if reflect.TypeOf(ev).Kind() != reflect.String {
		return "", errors.New("value is not string")
	} else if ev.(string) == "[NOT EVALUABLE]" {
		return "", errors.New("value is not evaluable")
	}

	return ev.(string), nil
}
