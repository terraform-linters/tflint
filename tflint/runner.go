package tflint

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/lang"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner checks templates according rules.
// For variables interplation, it has Terraform eval context,
// and state. After checking, it accumulates results as issues.
type Runner struct {
	TFConfig *configs.Config
	Issues   issue.Issues

	ctx    terraform.BuiltinEvalContext
	state  state.TFState
	config *config.Config
	logger *logger.Logger
}

// NewRunner returns new TFLint runner
// TODO: Generate variables from configs
func NewRunner(cfg *configs.Config) *Runner {
	return &Runner{
		ctx: terraform.BuiltinEvalContext{
			Evaluator: &terraform.Evaluator{
				Config:             cfg,
				VariableValues:     map[string]map[string]cty.Value{},
				VariableValuesLock: &sync.Mutex{},
			},
		},
		TFConfig: cfg,
		Issues:   []*issue.Issue{},
	}
}

// EvaluateExpr is a wrapper of terraform.BultinEvalContext.EvaluateExpr and gocty.FromCtyValue
// When it received slice as `ret`, it converts cty.Value to expected list type
// because raw cty.Value has TupleType.
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	val, diags := r.ctx.EvaluateExpr(expr, cty.DynamicPseudoType, nil)
	if diags.HasErrors() {
		return &Error{
			Code: EvaluationError,
			Message: fmt.Sprintf(
				"Failed to eval an expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: diags.Err(),
		}
	}

	if !val.IsKnown() {
		return &Error{
			Code: UnknownValueError,
			Message: fmt.Sprintf(
				"Unknown value found in %s:%d; Please use environment variables or tfvars to set the value",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
		}
	}

	var err error
	switch ret.(type) {
	case *string:
		val, err = convert.Convert(val, cty.String)
	case *int:
		val, err = convert.Convert(val, cty.Number)
	case *[]string:
		val, err = convert.Convert(val, cty.List(cty.String))
	case *[]int:
		val, err = convert.Convert(val, cty.List(cty.Number))
	}

	if err != nil {
		return &Error{
			Code: TypeConversionError,
			Message: fmt.Sprintf(
				"Invalid type expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: err,
		}
	}

	err = gocty.FromCtyValue(val, ret)
	if err != nil {
		return &Error{
			Code: TypeMismatchError,
			Message: fmt.Sprintf(
				"Invalid type expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: err,
		}
	}
	return nil
}

// TODO: Move to EvaluateExpr
func isEvaluable(expr hcl.Expression) bool {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		panic(diags.Err())
	}
	for _, ref := range refs {
		switch ref.Subject.(type) {
		case addrs.InputVariable:
			// noop
		case addrs.TerraformAttr:
			// noop
		default:
			return false
		}
	}
	return true
}

func getWorkspace() string {
	if envVar := os.Getenv("TF_WORKSPACE"); envVar != "" {
		return envVar
	}

	envData, _ := ioutil.ReadFile(".terraform/environment")
	current := string(bytes.TrimSpace(envData))
	if current == "" {
		current = "default"
	}

	return current
}
