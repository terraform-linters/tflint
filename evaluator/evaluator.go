package evaluator

import (
	"strings"

	hcl_ast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hil"
	hil_ast "github.com/hashicorp/hil/ast"
)

type Evaluator struct {
	Config hil.EvalConfig
}

func NewEvaluator(listmap map[string]*hcl_ast.ObjectList) *Evaluator {
	varmap := detectVariables(listmap)

	evaluator := &Evaluator{
		Config: hil.EvalConfig{
			GlobalScope: &hil_ast.BasicScope{
				VarMap: varmap,
			},
		},
	}

	return evaluator
}

func detectVariables(listmap map[string]*hcl_ast.ObjectList) map[string]hil_ast.Variable {
	varmap := make(map[string]hil_ast.Variable)

	for _, list := range listmap {
		for _, item := range list.Filter("variable").Items {
			var variable hil_ast.Variable
			varName := "var." + strings.Trim(item.Keys[0].Token.Text, "\"")
			varTypeString := strings.Trim(item.Val.(*hcl_ast.ObjectType).List.Filter("type").Items[0].Val.(*hcl_ast.LiteralType).Token.Text, "\"")

			switch varTypeString {
			case "string":
				variable = hil_ast.Variable{
					Type:  hil_ast.TypeString,
					Value: strings.Trim(item.Val.(*hcl_ast.ObjectType).List.Filter("default").Items[0].Val.(*hcl_ast.LiteralType).Token.Text, "\""),
				}
			}
			varmap[varName] = variable
		}
	}

	return varmap
}

func (e *Evaluator) Eval(src string) string {
	root, _ := hil.Parse(src)
	result, _ := hil.Eval(root, &e.Config)
	return result.Value.(string)
}
