package detector

import (
	"errors"
	"fmt"
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

func Detect(listMap map[string]*ast.ObjectList) ([]*issue.Issue, error) {
	evalConfig, err := eval.NewEvaluator(listMap)
	if err != nil {
		return nil, err
	}

	detector := &Detector{
		ListMap:    listMap,
		EvalConfig: evalConfig,
	}
	return detector.detect(), nil
}

func hclLiteralToken(item *ast.ObjectItem, k string) (token.Token, error) {
	items := item.Val.(*ast.ObjectType).List.Filter(k).Items
	if len(items) == 0 {
		return token.Token{}, errors.New(fmt.Sprintf("ERROR: key `%s` not found", k))
	}

	if v, ok := items[0].Val.(*ast.LiteralType); ok {
		return v.Token, nil
	}
	return token.Token{}, errors.New(fmt.Sprintf("ERROR: `%s` value is not literal", k))
}

func IsKeyNotFound(item *ast.ObjectItem, k string) bool {
	items := item.Val.(*ast.ObjectType).List.Filter(k).Items
	return len(items) == 0
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
		return "", errors.New(fmt.Sprintf("ERROR: `%s` is not string", v))
	} else if ev.(string) == "[NOT EVALUABLE]" {
		return "", errors.New(fmt.Sprintf("ERROR; `%s` is not evaluable", v))
	}

	return ev.(string), nil
}
