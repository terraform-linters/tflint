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
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateTestDetector() *TestDetector {
	return &TestDetector{
		Detector:  d,
		IssueType: "TEST",
		Target:    "aws_instance",
		DeepCheck: false,
	}
}

func (d *TestDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	*issues = append(*issues, &issue.Issue{
		Type:    d.IssueType,
		Message: "this is test method",
		Line:    1,
		File:    file,
	})
}

func TestDetectByCreatorName(creatorMethod string, src string, stateJSON string, c *config.Config, awsClient *config.AwsClient, issues *[]*issue.Issue) error {
	templates := make(map[string]*ast.File)
	templates["test.tf"], _ = parser.Parse([]byte(src))

	tfstate := &state.TFState{}
	if err := json.Unmarshal([]byte(stateJSON), tfstate); err != nil && stateJSON != "" {
		return errors.New("Invalid JSON Syntax!")
	}

	evalConfig, _ := evaluator.NewEvaluator(templates, []*ast.File{}, c)
	creator := reflect.ValueOf(&Detector{
		Templates:  templates,
		State:      tfstate,
		EvalConfig: evalConfig,
		Config:     c,
		Logger:     logger.Init(false),
		AwsClient:  awsClient,
	}).MethodByName(creatorMethod)
	detector := creator.Call([]reflect.Value{})[0]

	if preprocess := detector.MethodByName("PreProcess"); preprocess.IsValid() {
		preprocess.Call([]reflect.Value{})
	}
	for file, template := range templates {
		for _, item := range template.Node.(*ast.ObjectList).Filter("resource", reflect.Indirect(detector).FieldByName("Target").String()).Items {
			detect := detector.MethodByName("Detect")
			detect.Call([]reflect.Value{
				reflect.ValueOf(file),
				reflect.ValueOf(item),
				reflect.ValueOf(issues),
			})
		}
	}
	return nil
}
