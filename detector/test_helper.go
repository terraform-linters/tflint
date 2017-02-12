package detector

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
)

type TestDetector struct {
	*Detector
}

func (d *Detector) CreateTestDetector() *TestDetector {
	return &TestDetector{d}
}

func (d *TestDetector) Detect(issues *[]*issue.Issue) {
	*issues = append(*issues, &issue.Issue{
		Type:    "TEST",
		Message: "this is test method",
		Line:    1,
		File:    "",
	})
}

func TestDetectByCreatorName(creatorMethod string, src string, stateJSON string, c *config.Config, awsClient *config.AwsClient, issues *[]*issue.Issue) error {
	listMap := make(map[string]*ast.ObjectList)
	root, _ := parser.Parse([]byte(src))
	list, _ := root.Node.(*ast.ObjectList)
	listMap["test.tf"] = list

	tfstate := &state.TFState{}
	if err := json.Unmarshal([]byte(stateJSON), tfstate); err != nil && stateJSON != "" {
		return errors.New("Invalid JSON Syntax!")
	}

	evalConfig, _ := evaluator.NewEvaluator(listMap, c)
	creator := reflect.ValueOf(&Detector{
		ListMap:       listMap,
		State:         tfstate,
		EvalConfig:    evalConfig,
		Config:        c,
		Logger:        logger.Init(false),
		AwsClient:     awsClient,
		ResponseCache: &ResponseCache{},
	}).MethodByName(creatorMethod)
	detector := creator.Call([]reflect.Value{})[0]

	method := detector.MethodByName("Detect")
	method.Call([]reflect.Value{reflect.ValueOf(issues)})
	return nil
}
