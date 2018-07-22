package tflint

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	logger *logger.Logger // TODO: Improve logging system
}

// NewRunner returns new TFLint runner
// TODO: Generate variables from configs
func NewRunner(cfg *configs.Config) *Runner {
	return &Runner{
		TFConfig: cfg,
		Issues:   []*issue.Issue{},

		ctx: terraform.BuiltinEvalContext{
			Evaluator: &terraform.Evaluator{
				Meta: &terraform.ContextMeta{
					Env: getWorkspace(),
				},
				Config:             cfg,
				VariableValues:     map[string]map[string]cty.Value{},
				VariableValuesLock: &sync.Mutex{},
			},
		},
		logger: logger.Init(false),
	}
}

// EvaluateExpr is a wrapper of terraform.BultinEvalContext.EvaluateExpr and gocty.FromCtyValue
// When it received slice as `ret`, it converts cty.Value to expected list type
// because raw cty.Value has TupleType.
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	if !isEvaluable(expr) {
		err := &Error{
			Code:  UnevaluableError,
			Level: WarningLevel,
			Message: fmt.Sprintf(
				"Unevaluable expression found in %s:%d",
				r.GetFileName(expr.Range().Filename),
				expr.Range().Start.Line,
			),
		}
		r.logger.Error(err)
		return err
	}

	val, diags := r.ctx.EvaluateExpr(expr, cty.DynamicPseudoType, nil)
	if diags.HasErrors() {
		err := &Error{
			Code:  EvaluationError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Failed to eval an expression in %s:%d",
				r.GetFileName(expr.Range().Filename),
				expr.Range().Start.Line,
			),
			Cause: diags.Err(),
		}
		r.logger.Error(err)
		return err
	}

	if !val.IsKnown() {
		err := &Error{
			Code:  UnknownValueError,
			Level: WarningLevel,
			Message: fmt.Sprintf(
				"Unknown value found in %s:%d; Please use environment variables or tfvars to set the value",
				r.GetFileName(expr.Range().Filename),
				expr.Range().Start.Line,
			),
		}
		r.logger.Error(err)
		return err
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
		err := &Error{
			Code:  TypeConversionError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Invalid type expression in %s:%d",
				r.GetFileName(expr.Range().Filename),
				expr.Range().Start.Line,
			),
			Cause: err,
		}
		r.logger.Error(err)
		return err
	}

	err = gocty.FromCtyValue(val, ret)
	if err != nil {
		err := &Error{
			Code:  TypeMismatchError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Invalid type expression in %s:%d",
				r.GetFileName(expr.Range().Filename),
				expr.Range().Start.Line,
			),
			Cause: err,
		}
		r.logger.Error(err)
		return err
	}
	return nil
}

// GetFileName returns user-friendly file name.
// It returns base file name when processing root module.
// Otherwise, it add the module name as prefix to base file name.
func (r *Runner) GetFileName(raw string) string {
	if r.TFConfig.Path.IsRoot() {
		return filepath.Base(raw)
	}
	return filepath.Join(r.TFConfig.Path.String(), filepath.Base(raw))
}

func isEvaluable(expr hcl.Expression) bool {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		// Maybe this is bug
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

	dir := os.Getenv("TF_DATA_DIR")
	if dir == "" {
		dir = ".terraform"
	}

	envData, _ := ioutil.ReadFile(filepath.Join(dir, "environment"))
	current := string(bytes.TrimSpace(envData))
	if current == "" {
		current = "default"
	}

	return current
}
