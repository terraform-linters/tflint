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
	"github.com/wata727/tflint/schema"
	"github.com/wata727/tflint/state"
)

type DetectorIF interface {
	Detect() []*issue.Issue
	HasError() bool
}

type Detector struct {
	Schema     []*schema.Template
	State      *state.TFState
	Config     *config.Config
	AwsClient  *config.AwsClient
	EvalConfig *evaluator.Evaluator
	Logger     *logger.Logger
	Error      bool

	Name       string
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
	Link       string
	Enabled    bool
}

var detectorFactories = []string{
	"CreateAwsInstanceInvalidTypeDetector",
	"CreateAwsInstancePreviousTypeDetector",
	"CreateAwsInstanceDefaultStandardVolumeDetector",
	"CreateAwsInstanceInvalidIAMProfileDetector",
	"CreateAwsInstanceInvalidAMIDetector",
	"CreateAwsInstanceInvalidKeyNameDetector",
	"CreateAwsInstanceInvalidSubnetDetector",
	"CreateAwsInstanceInvalidVPCSecurityGroupDetector",
	"CreateAwsALBInvalidSecurityGroupDetector",
	"CreateAwsALBInvalidSubnetDetector",
	"CreateAwsALBDuplicateNameDetector",
	"CreateAwsELBInvalidSecurityGroupDetector",
	"CreateAwsELBInvalidSubnetDetector",
	"CreateAwsELBInvalidInstanceDetector",
	"CreateAwsELBDuplicateNameDetector",
	"CreateAwsDBInstanceDefaultParameterGroupDetector",
	"CreateAwsDBInstanceInvalidVPCSecurityGroupDetector",
	"CreateAwsDBInstanceInvalidDBSubnetGroupDetector",
	"CreateAwsDBInstanceInvalidParameterGroupDetector",
	"CreateAwsDBInstanceInvalidOptionGroupDetector",
	"CreateAwsDBInstanceInvalidTypeDetector",
	"CreateAwsDBInstancePreviousTypeDetector",
	"CreateAwsDBInstanceReadablePasswordDetector",
	"CreateAwsDBInstanceDuplicateIdentifierDetector",
	"CreateAwsElastiCacheClusterDefaultParameterGroupDetector",
	"CreateAwsElastiCacheClusterInvalidParameterGroupDetector",
	"CreateAwsElastiCacheClusterInvalidSubnetGroupDetector",
	"CreateAwsElastiCacheClusterInvalidSecurityGroupDetector",
	"CreateAwsElastiCacheClusterInvalidTypeDetector",
	"CreateAwsElastiCacheClusterPreviousTypeDetector",
	"CreateAwsElastiCacheClusterDuplicateIDDetector",
	"CreateAwsSecurityGroupDuplicateDetector",
	"CreateAwsRouteInvalidRouteTableDetector",
	"CreateAwsRouteNotSpecifiedTargetDetector",
	"CreateAwsRouteSpecifiedMultipleTargetsDetector",
	"CreateAwsRouteInvalidGatewayDetector",
	"CreateAwsRouteInvalidEgressOnlyGatewayDetector",
	"CreateAwsRouteInvalidNatGatewayDetector",
	"CreateAwsRouteInvalidVpcPeeringConnectionDetector",
	"CreateAwsRouteInvalidInstanceDetector",
	"CreateAwsRouteInvalidNetworkInterfaceDetector",
	"CreateAwsCloudWatchMetricAlarmInvalidUnitDetector",
	"CreateAwsECSClusterDuplicateNameDetector",
	"CreateTerraformModulePinnedSourceDetector",
}

func NewDetector(templates map[string]*ast.File, schema []*schema.Template, state *state.TFState, tfvars []*ast.File, c *config.Config) (*Detector, error) {
	evalConfig, err := evaluator.NewEvaluator(templates, schema, tfvars, c)
	if err != nil {
		return nil, err
	}

	return &Detector{
		Schema:     schema,
		State:      state,
		Config:     c,
		AwsClient:  c.NewAwsClient(),
		EvalConfig: evalConfig,
		Logger:     logger.Init(c.Debug),
		Error:      false,
	}, nil
}

func (d *Detector) Detect() []*issue.Issue {
	var issues = []*issue.Issue{}
	for _, creatorMethod := range detectorFactories {
		d.detect(creatorMethod, &issues)

		for _, template := range d.Schema {
			for _, module := range template.Modules {
				if d.Config.IgnoreModule[module.ModuleSource] {
					d.Logger.Info(fmt.Sprintf("ignore module `%s`", module.Id))
					continue
				}
				d.Logger.Info(fmt.Sprintf("detect module `%s`", module.Id))
				moduleDetector, err := NewDetector(map[string]*ast.File{}, module.Templates, d.State, []*ast.File{}, d.Config)
				if err != nil {
					d.Logger.Error(err)
					continue
				}
				moduleDetector.EvalConfig = &evaluator.Evaluator{
					Config: module.EvalConfig,
				}
				moduleDetector.Error = d.Error
				moduleDetector.detect(creatorMethod, &issues)
			}
		}
	}

	return issues
}

func (d *Detector) HasError() bool {
	return d.Error
}

func (d *Detector) detect(creatorMethod string, issues *[]*issue.Issue) {
	creator := reflect.ValueOf(d).MethodByName(creatorMethod)
	detector := creator.Call([]reflect.Value{})[0]
	ruleName := reflect.Indirect(detector).FieldByName("Name").String()

	if d.isSkip(
		ruleName,
		reflect.Indirect(detector).FieldByName("Enabled").Bool(),
		reflect.Indirect(detector).FieldByName("DeepCheck").Bool(),
		reflect.Indirect(detector).FieldByName("TargetType").String(),
		reflect.Indirect(detector).FieldByName("Target").String(),
	) {
		d.Logger.Info(fmt.Sprintf("skip `%s`", ruleName))
		return
	}

	d.Logger.Info(fmt.Sprintf("detect by `%s`", ruleName))
	if preProcess := detector.MethodByName("PreProcess"); preProcess.IsValid() {
		preProcess.Call([]reflect.Value{})
	}

	switch reflect.Indirect(detector).FieldByName("TargetType").String() {
	case "resource":
		for _, template := range d.Schema {
			for _, resource := range template.FindResources(reflect.Indirect(detector).FieldByName("Target").String()) {
				detect := detector.MethodByName("Detect")
				detect.Call([]reflect.Value{
					reflect.ValueOf(resource),
					reflect.ValueOf(issues),
				})
			}
		}
	case "module":
		for _, template := range d.Schema {
			for _, module := range template.Modules {
				detect := detector.MethodByName("Detect")
				detect.Call([]reflect.Value{
					reflect.ValueOf(module),
					reflect.ValueOf(issues),
				})
			}
		}
	default:
		d.Logger.Info("Unexpected target type.")
		return
	}
}

func (d *Detector) evalToString(v string) (string, error) {
	ev, err := d.EvalConfig.Eval(strings.Trim(v, "\""))

	if err != nil {
		return "", err
	} else if reflect.TypeOf(ev).Kind() != reflect.String {
		return "", fmt.Errorf("ERROR: `%s` is not string", v)
	} else if ev.(string) == "[NOT EVALUABLE]" {
		return "", fmt.Errorf("ERROR: `%s` is not evaluable", v)
	}

	return ev.(string), nil
}

func (d *Detector) evalToStringTokens(t token.Token) ([]token.Token, error) {
	ev, err := d.EvalConfig.Eval(strings.Trim(t.Text, "\""))

	if err != nil {
		return []token.Token{}, err
	} else if reflect.TypeOf(ev).Kind() != reflect.Slice {
		return []token.Token{}, fmt.Errorf("ERROR: `%s` is not list", t.Text)
	} else if sev, _ := ev.(string); sev == "[NOT EVALUABLE]" {
		return []token.Token{}, fmt.Errorf("ERROR: `%s` is not evaluable", t.Text)
	}

	var tokens []token.Token
	for _, node := range ev.([]interface{}) {
		if reflect.TypeOf(node).Kind() != reflect.String {
			return []token.Token{}, fmt.Errorf("ERROR: `%s` contains not string value", t.Text)
		}
		nodeToken := token.Token{
			Text: node.(string),
			Pos: token.Pos{
				Line: t.Pos.Line,
			},
		}
		tokens = append(tokens, nodeToken)
	}

	return tokens, nil
}

func (d *Detector) isSkip(name string, enabled bool, deepCheck bool, targetType string, target string) bool {
	if rule := d.Config.Rules[name]; rule != nil {
		if rule.Enabled {
			// noop
		} else if !rule.Enabled || !enabled {
			return true
		}
	} else {
		if d.Config.IgnoreRule[name] || !enabled {
			return true
		}
	}

	if deepCheck && !d.Config.DeepCheck {
		return true
	}

	targets := 0
	switch targetType {
	case "resource":
		for _, template := range d.Schema {
			targets += len(template.FindResources(target))
		}
	case "module":
		for _, template := range d.Schema {
			targets += len(template.Modules)
		}
	default:
		d.Logger.Info("Unexpected target type.")
	}

	if targets == 0 {
		d.Logger.Info("targets are not found.")
		return true
	}
	return false
}
