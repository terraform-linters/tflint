package evaluator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hil_ast "github.com/hashicorp/hil/ast"
)

type Evaluator struct {
	Config       hil.EvalConfig
	ModuleConfig map[string]*hclModule
}

func NewEvaluator(listMap map[string]*hcl_ast.ObjectList) (*Evaluator, error) {
	varMap, err := detectVariables(listMap)
	if err != nil {
		return nil, err
	}
	moduleMap, err := detectModules(listMap)
	if err != nil {
		return nil, err
	}

	evaluator := &Evaluator{
		Config: hil.EvalConfig{
			GlobalScope: &hil_ast.BasicScope{
				VarMap: varMap,
			},
		},
		ModuleConfig: moduleMap,
	}

	return evaluator, nil
}

func isEvaluable(src string) bool {
	var supportPrefix = []string{
		"var",
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
		return nil, errors.New(fmt.Sprintf("ERROR: unexcepted type variable `%s`", src))
	}
}
