package detector

import (
	"reflect"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
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

func TestDetectByCreatorName(creatorMethod string, src string, c *config.Config, awsClient *config.AwsClient, issues *[]*issue.Issue) {
	listMap := make(map[string]*ast.ObjectList)
	root, _ := parser.Parse([]byte(src))
	list, _ := root.Node.(*ast.ObjectList)
	listMap["test.tf"] = list

	evalConfig, _ := evaluator.NewEvaluator(listMap, c)
	creator := reflect.ValueOf(&Detector{
		ListMap:    listMap,
		EvalConfig: evalConfig,
		Config:     c,
		AwsClient:  awsClient,
	}).MethodByName(creatorMethod)
	detector := creator.Call([]reflect.Value{})[0]

	method := detector.MethodByName("Detect")
	method.Call([]reflect.Value{reflect.ValueOf(issues)})
}
