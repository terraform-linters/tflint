package detector

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/logger"
)

type Detector struct {
	ListMap    map[string]*ast.ObjectList
	Config     *config.Config
	EvalConfig *evaluator.Evaluator
	Logger     *logger.Logger
}

var detectors = map[string]string{
	"aws_instance_invalid_type":                       "DetectAwsInstanceInvalidType",
	"aws_instance_previous_type":                      "DetectAwsInstancePreviousType",
	"aws_instance_not_specified_iam_profile":          "DetectAwsInstanceNotSpecifiedIamProfile",
	"aws_instance_default_standard_volume":            "DetectAwsInstanceDefaultStandardVolume",
	"aws_db_instance_default_parameter_group":         "DetectAwsDbInstanceDefaultParameterGroup",
	"aws_elasticache_cluster_default_parameter_group": "DetectAwsElasticacheClusterDefaultParameterGroup",
}

func NewDetector(listMap map[string]*ast.ObjectList, c *config.Config) (*Detector, error) {
	evalConfig, err := evaluator.NewEvaluator(listMap, c)
	if err != nil {
		return nil, err
	}

	return &Detector{
		ListMap:    listMap,
		Config:     c,
		EvalConfig: evalConfig,
		Logger:     logger.Init(c.Debug),
	}, nil
}

func hclLiteralToken(item *ast.ObjectItem, k string) (token.Token, error) {
	objItems, err := hclObjectItems(item, k)
	if err != nil {
		return token.Token{}, err
	}

	if v, ok := objItems[0].Val.(*ast.LiteralType); ok {
		return v.Token, nil
	}
	return token.Token{}, fmt.Errorf("ERROR: `%s` value is not literal", k)
}

func hclObjectItems(item *ast.ObjectItem, k string) ([]*ast.ObjectItem, error) {
	items := item.Val.(*ast.ObjectType).List.Filter(k).Items
	if len(items) == 0 {
		return []*ast.ObjectItem{}, fmt.Errorf("ERROR: key `%s` not found", k)
	}
	return items, nil
}

func IsKeyNotFound(item *ast.ObjectItem, k string) bool {
	items := item.Val.(*ast.ObjectType).List.Filter(k).Items
	return len(items) == 0
}

func (d *Detector) Detect() []*issue.Issue {
	var issues = []*issue.Issue{}

	for ruleName, detectorMethod := range detectors {
		if d.Config.IgnoreRule[ruleName] {
			d.Logger.Info(fmt.Sprintf("ignore rule `%s`", ruleName))
			continue
		}
		d.Logger.Info(fmt.Sprintf("detect by `%s`", ruleName))
		method := reflect.ValueOf(d).MethodByName(detectorMethod)
		method.Call([]reflect.Value{reflect.ValueOf(&issues)})

		for name, m := range d.EvalConfig.ModuleConfig {
			if d.Config.IgnoreModule[m.Source] {
				d.Logger.Info(fmt.Sprintf("ignore module `%s`", name))
				continue
			}
			d.Logger.Info(fmt.Sprintf("detect module `%s`", name))
			moduleDetector := &Detector{
				ListMap: m.ListMap,
				Config:  d.Config,
				EvalConfig: &evaluator.Evaluator{
					Config: m.Config,
				},
				Logger: d.Logger,
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
		return "", fmt.Errorf("ERROR: `%s` is not string", v)
	} else if ev.(string) == "[NOT EVALUABLE]" {
		return "", fmt.Errorf("ERROR; `%s` is not evaluable", v)
	}

	return ev.(string), nil
}
