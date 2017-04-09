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
	"github.com/wata727/tflint/state"
)

type DetectorIF interface {
	Detect() []*issue.Issue
	HasError() bool
}

type Detector struct {
	Templates  map[string]*ast.File
	State      *state.TFState
	Config     *config.Config
	AwsClient  *config.AwsClient
	EvalConfig *evaluator.Evaluator
	Logger     *logger.Logger
	Error      bool
}

var detectors = map[string]string{
	"aws_instance_invalid_type":                       "CreateAwsInstanceInvalidTypeDetector",
	"aws_instance_previous_type":                      "CreateAwsInstancePreviousTypeDetector",
	"aws_instance_not_specified_iam_profile":          "CreateAwsInstanceNotSpecifiedIAMProfileDetector",
	"aws_instance_default_standard_volume":            "CreateAwsInstanceDefaultStandardVolumeDetector",
	"aws_instance_invalid_iam_profile":                "CreateAwsInstanceInvalidIAMProfileDetector",
	"aws_instance_invalid_ami":                        "CreateAwsInstanceInvalidAMIDetector",
	"aws_instance_invalid_key_name":                   "CreateAwsInstanceInvalidKeyNameDetector",
	"aws_instance_invalid_subnet":                     "CreateAwsInstanceInvalidSubnetDetector",
	"aws_instance_invalid_vpc_security_group":         "CreateAwsInstanceInvalidVPCSecurityGroupDetector",
	"aws_alb_invalid_security_group":                  "CreateAwsALBInvalidSecurityGroupDetector",
	"aws_alb_invalid_subnet":                          "CreateAwsALBInvalidSubnetDetector",
	"aws_alb_duplicate_name":                          "CreateAwsALBDuplicateNameDetector",
	"aws_elb_invalid_security_group":                  "CreateAwsELBInvalidSecurityGroupDetector",
	"aws_elb_invalid_subnet":                          "CreateAwsELBInvalidSubnetDetector",
	"aws_elb_invalid_instance":                        "CreateAwsELBInvalidInstanceDetector",
	"aws_elb_duplicate_name":                          "CreateAwsELBDuplicateNameDetector",
	"aws_db_instance_default_parameter_group":         "CreateAwsDBInstanceDefaultParameterGroupDetector",
	"aws_db_instance_invalid_vpc_security_group":      "CreateAwsDBInstanceInvalidVPCSecurityGroupDetector",
	"aws_db_instance_invalid_db_subnet_group":         "CreateAwsDBInstanceInvalidDBSubnetGroupDetector",
	"aws_db_instance_invalid_parameter_group":         "CreateAwsDBInstanceInvalidParameterGroupDetector",
	"aws_db_instance_invalid_option_group":            "CreateAwsDBInstanceInvalidOptionGroupDetector",
	"aws_db_instance_invalid_type":                    "CreateAwsDBInstanceInvalidTypeDetector",
	"aws_db_instance_previous_type":                   "CreateAwsDBInstancePreviousTypeDetector",
	"aws_db_instance_readable_password":               "CreateAwsDBInstanceReadablePasswordDetector",
	"aws_db_instance_duplicate_identifier":            "CreateAwsDBInstanceDuplicateIdentifierDetector",
	"aws_elasticache_cluster_default_parameter_group": "CreateAwsElastiCacheClusterDefaultParameterGroupDetector",
	"aws_elasticache_cluster_invalid_parameter_group": "CreateAwsElastiCacheClusterInvalidParameterGroupDetector",
	"aws_elasticache_cluster_invalid_subnet_group":    "CreateAwsElastiCacheClusterInvalidSubnetGroupDetector",
	"aws_elasticache_cluster_invalid_security_group":  "CreateAwsElastiCacheClusterInvalidSecurityGroupDetector",
	"aws_elasticache_cluster_invalid_type":            "CreateAwsElastiCacheClusterInvalidTypeDetector",
	"aws_elasticache_cluster_previous_type":           "CreateAwsElastiCacheClusterPreviousTypeDetector",
	"aws_elasticache_cluster_duplicate_id":            "CreateAwsElastiCacheClusterDuplicateIDDetector",
	"aws_security_group_duplicate_name":               "CreateAwsSecurityGroupDuplicateDetector",
	"aws_route_invalid_route_table":                   "CreateAwsRouteInvalidRouteTableDetector",
	"aws_route_not_specified_target":                  "CreateAwsRouteNotSpecifiedTargetDetector",
	"aws_route_specified_multiple_targets":            "CreateAwsRouteSpecifiedMultipleTargetsDetector",
	"aws_route_invalid_gateway":                       "CreateAwsRouteInvalidGatewayDetector",
}

func NewDetector(templates map[string]*ast.File, state *state.TFState, tfvars []*ast.File, c *config.Config) (*Detector, error) {
	evalConfig, err := evaluator.NewEvaluator(templates, tfvars, c)
	if err != nil {
		return nil, err
	}

	return &Detector{
		Templates:  templates,
		State:      state,
		Config:     c,
		AwsClient:  c.NewAwsClient(),
		EvalConfig: evalConfig,
		Logger:     logger.Init(c.Debug),
		Error:      false,
	}, nil
}

func hclObjectKeyText(item *ast.ObjectItem) string {
	return strings.Trim(item.Keys[0].Token.Text, "\"")
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

func hclLiteralListToken(item *ast.ObjectItem, k string) ([]token.Token, error) {
	objItems, err := hclObjectItems(item, k)
	if err != nil {
		return []token.Token{}, err
	}

	var tokens []token.Token
	if v, ok := objItems[0].Val.(*ast.ListType); ok {
		for _, node := range v.List {
			if v, ok := node.(*ast.LiteralType); ok {
				tokens = append(tokens, v.Token)
			} else {
				return []token.Token{}, fmt.Errorf("ERROR: `%s` contains not literal value", k)
			}
		}
		return tokens, nil
	}
	return []token.Token{}, fmt.Errorf("ERROR: `%s` value is not list", k)
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
	modules := d.EvalConfig.ModuleConfig

	for ruleName, creatorMethod := range detectors {
		if d.Config.IgnoreRule[ruleName] {
			d.Logger.Info(fmt.Sprintf("ignore rule `%s`", ruleName))
			continue
		}
		d.Logger.Info(fmt.Sprintf("detect by `%s`", ruleName))
		d.detect(creatorMethod, &issues)

		for name, m := range modules {
			if d.Config.IgnoreModule[m.Source] {
				d.Logger.Info(fmt.Sprintf("ignore module `%s`", name))
				continue
			}
			d.Logger.Info(fmt.Sprintf("detect module `%s`", name))
			moduleDetector, err := NewDetector(m.Templates, d.State, []*ast.File{}, d.Config)
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			moduleDetector.EvalConfig = &evaluator.Evaluator{
				Config: m.Config,
			}
			moduleDetector.Error = d.Error
			moduleDetector.detect(creatorMethod, &issues)
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

	if d.isSkip(reflect.Indirect(detector).FieldByName("DeepCheck").Bool(), reflect.Indirect(detector).FieldByName("Target").String()) {
		d.Logger.Info("skip this rule.")
		return
	}
	if preprocess := detector.MethodByName("PreProcess"); preprocess.IsValid() {
		preprocess.Call([]reflect.Value{})
	}
	for file, template := range d.Templates {
		for _, item := range template.Node.(*ast.ObjectList).Filter("resource", reflect.Indirect(detector).FieldByName("Target").String()).Items {
			detect := detector.MethodByName("Detect")
			detect.Call([]reflect.Value{
				reflect.ValueOf(file),
				reflect.ValueOf(item),
				reflect.ValueOf(issues),
			})
		}
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

func (d *Detector) isSkip(deepCheck bool, target string) bool {
	if deepCheck && !d.Config.DeepCheck {
		return true
	}

	targetResources := 0
	for _, template := range d.Templates {
		targetResources += len(template.Node.(*ast.ObjectList).Filter("resource", target).Items)
	}
	if targetResources == 0 {
		d.Logger.Info("target resources are not found.")
		return true
	}
	return false
}
