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
	"github.com/wata727/tflint/schema"
	"github.com/wata727/tflint/state"
)

type TestDetector struct {
	*Detector
}

func (d *Detector) CreateTestDetector() *TestDetector {
	nd := &TestDetector{Detector: d}
	nd.Name = "test_rule"
	nd.IssueType = "TEST"
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = false
	nd.Enabled = true
	return nd
}

func (d *TestDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	*issues = append(*issues, &issue.Issue{
		Type:    d.IssueType,
		Message: "this is test method",
		Line:    1,
		File:    resource.File,
	})
}

func TestDetectByCreatorName(creatorMethod string, src string, stateJSON string, c *config.Config, awsClient *config.AwsClient, issues *[]*issue.Issue, filename string) error {
	templates := make(map[string]*ast.File)
	filename_to_use := filename || "test.tf"
	templates[filename_to_use], _ = parser.Parse([]byte(src))
	files := map[string][]byte{filename_to_use: []byte(src)}

	tfstate := &state.TFState{}
	if err := json.Unmarshal([]byte(stateJSON), tfstate); err != nil && stateJSON != "" {
		return errors.New("Invalid JSON Syntax!")
	}
	schema, _ := schema.Make(files)

	evalConfig, _ := evaluator.NewEvaluator(templates, schema, []*ast.File{}, c)
	creator := reflect.ValueOf(&Detector{
		Schema:     schema,
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
	switch reflect.Indirect(detector).FieldByName("TargetType").String() {
	case "resource":
		for _, template := range schema {
			for _, resource := range template.FindResources(reflect.Indirect(detector).FieldByName("Target").String()) {
				detect := detector.MethodByName("Detect")
				detect.Call([]reflect.Value{
					reflect.ValueOf(resource),
					reflect.ValueOf(issues),
				})
			}
		}
	case "provider":
		for _, template := range schema {
			for _, provider := range template.FindProviders(reflect.Indirect(detector).FieldByName("Target").String()) {
				detect := detector.MethodByName("Detect")
				detect.Call([]reflect.Value{
					reflect.ValueOf(provider),
					reflect.ValueOf(issues),
				})
			}
		}
	case "module":
		for _, template := range schema {
			for _, module := range template.Modules {
				detect := detector.MethodByName("Detect")
				detect.Call([]reflect.Value{
					reflect.ValueOf(module),
					reflect.ValueOf(issues),
				})
			}
		}
	default:
		return nil
	}
	return nil
}
