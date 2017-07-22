package evaluator

import (
	"fmt"
	"regexp"
	"strings"

	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/schema"
)

type Evaluator struct {
	Config hil.EvalConfig
}

func NewEvaluator(templates map[string]*hclast.File, schema []*schema.Template, varfile []*hclast.File, c *config.Config) (*Evaluator, error) {
	varMap, err := detectVariables(templates, varfile)
	if err != nil {
		return nil, err
	}
	varMap["terraform.env"] = hilast.Variable{
		Type:  hilast.TypeString,
		Value: c.TerraformEnv,
	}

	evaluator := &Evaluator{
		Config: hil.EvalConfig{
			GlobalScope: &hilast.BasicScope{
				VarMap: varMap,
			},
		},
	}

	for _, template := range schema {
		for _, module := range template.Modules {
			if err := evaluator.initModule(module, c); err != nil {
				return nil, err
			}
		}
	}

	return evaluator, nil
}

func isEvaluable(src string) bool {
	var supportPrefix = []string{
		"var",
		"terraform",
	}

	supportPrefixPattern := "("
	for _, v := range supportPrefix {
		supportPrefixPattern = supportPrefixPattern + v + "|"
	}
	supportPrefixPattern = strings.Trim(supportPrefixPattern, "|") + ")"
	supportHilPattern := "\\${" + supportPrefixPattern + "\\..+}"

	return regexp.MustCompile(supportHilPattern).Match([]byte(src)) || !strings.Contains(src, "$")
}

func (e *Evaluator) Eval(src string) (interface{}, error) {
	if !isEvaluable(src) {
		return "[NOT EVALUABLE]", nil
	}
	root, err := hil.Parse(src)
	if err != nil {
		return nil, err
	}
	result, err := hil.Eval(root, &e.Config)
	if err != nil {
		return nil, err
	}

	switch result.Type.String() {
	case "TypeString":
		return result.Value.(string), nil
	case "TypeList":
		return result.Value.([]interface{}), nil
	case "TypeMap":
		return result.Value.(map[string]interface{}), nil
	case "TypeInt":
		return result.Value.(int), nil
	default:
		return nil, fmt.Errorf("ERROR: unexcepted type variable `%s`", src)
	}
}
