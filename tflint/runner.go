package tflint

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/lang"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/issue"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Runner checks templates according rules.
// For variables interplation, it has Terraform eval context.
// After checking, it accumulates results as issues.
type Runner struct {
	TFConfig  *configs.Config
	Issues    issue.Issues
	AwsClient *client.AwsClient

	ctx    terraform.BuiltinEvalContext
	config *Config
}

// NewRunner returns new TFLint runner
// It prepares built-in context (workpace metadata, variables) from
// received `configs.Config` and `terraform.InputValues`
func NewRunner(c *Config, cfg *configs.Config, variables ...terraform.InputValues) *Runner {
	path := "root"
	if !cfg.Path.IsRoot() {
		path = cfg.Path.String()
	}
	log.Printf("[INFO] Initialize new runner for %s", path)

	return &Runner{
		TFConfig:  cfg,
		Issues:    []*issue.Issue{},
		AwsClient: client.NewAwsClient(c.AwsCredentials),

		ctx: terraform.BuiltinEvalContext{
			Evaluator: &terraform.Evaluator{
				Meta: &terraform.ContextMeta{
					Env: getTFWorkspace(),
				},
				Config:             cfg,
				VariableValues:     prepareVariableValues(cfg.Module.Variables, variables...),
				VariableValuesLock: &sync.Mutex{},
			},
		},
		config: c,
	}
}

// NewModuleRunners returns new TFLint runners for child modules
// Recursively search modules and generate Runners
// In order to propagate attributes of moduleCall as variables to the module,
// evaluate the variables. If it cannot be evaluated, treat it as unknown
func NewModuleRunners(parent *Runner) ([]*Runner, error) {
	runners := []*Runner{}

	for name, cfg := range parent.TFConfig.Children {
		moduleCall, ok := parent.TFConfig.Module.ModuleCalls[name]
		if !ok {
			panic(fmt.Errorf("Expected module call `%s` is not found in `%s`", name, parent.TFConfig.Path.String()))
		}
		if parent.TFConfig.Path.IsRoot() && parent.config.IgnoreModule[moduleCall.SourceAddr] {
			log.Printf("[INFO] Ignore `%s` module", moduleCall.Name)
			continue
		}

		attributes, diags := moduleCall.Config.JustAttributes()
		if diags.HasErrors() {
			var causeErr error
			if diags[0].Subject == nil {
				// HACK: When Subject is nil, it outputs unintended message, so it replaces with actual file.
				causeErr = errors.New(strings.Replace(diags.Error(), "<nil>: ", "", 1))
			} else {
				causeErr = diags
			}
			err := &Error{
				Code:  UnexpectedAttributeError,
				Level: ErrorLevel,
				Message: fmt.Sprintf(
					"Attribute of module not allowed was found in %s:%d",
					parent.GetFileName(moduleCall.DeclRange.Filename),
					moduleCall.DeclRange.Start.Line,
				),
				Cause: causeErr,
			}
			log.Printf("[ERROR] %s", err)
			return runners, err
		}

		for varName, rawVar := range cfg.Module.Variables {
			if attribute, exists := attributes[varName]; exists {
				if isEvaluable(attribute.Expr) {
					val, diags := parent.ctx.EvaluateExpr(attribute.Expr, cty.DynamicPseudoType, nil)
					if diags.HasErrors() {
						err := &Error{
							Code:  EvaluationError,
							Level: ErrorLevel,
							Message: fmt.Sprintf(
								"Failed to eval an expression in %s:%d",
								parent.GetFileName(attribute.Expr.Range().Filename),
								attribute.Expr.Range().Start.Line,
							),
							Cause: diags.Err(),
						}
						log.Printf("[ERROR] %s", err)
						return runners, err
					}
					rawVar.Default = val
				} else {
					// If module attributes are not evaluable, it marks that value as unknown.
					// Unknown values are ignored when evaluated inside the module.
					log.Printf("[DEBUG] `%s` has been marked as unknown", varName)
					rawVar.Default = cty.UnknownVal(cty.DynamicPseudoType)
				}
			}
		}

		runner := NewRunner(parent.config, cfg)
		runners = append(runners, runner)
		moudleRunners, err := NewModuleRunners(runner)
		if err != nil {
			return runners, err
		}
		runners = append(runners, moudleRunners...)
	}

	return runners, nil
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
		log.Printf("[WARN] %s; TFLint ignores an unevaluable expression.", err)
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
		log.Printf("[ERROR] %s", err)
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
		log.Printf("[WARN] %s; TFLint ignores an expression includes an unknown value.", err)
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
	case *map[string]string:
		val, err = convert.Convert(val, cty.Map(cty.String))
	case *map[string]int:
		val, err = convert.Convert(val, cty.Map(cty.Number))
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
		log.Printf("[ERROR] %s", err)
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
		log.Printf("[ERROR] %s", err)
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

// TFConfigPath is a wrapper of addrs.Module
func (r *Runner) TFConfigPath() string {
	if r.TFConfig.Path.IsRoot() {
		return "root"
	}
	return r.TFConfig.Path.String()
}

// LookupIssues returns issues according to the received files
func (r *Runner) LookupIssues(files ...string) issue.Issues {
	if len(files) == 0 {
		return r.Issues
	}

	issues := []*issue.Issue{}
	for _, issue := range r.Issues {
		for _, file := range files {
			if file == issue.File {
				issues = append(issues, issue)
			}
		}
	}
	return issues
}

// WalkResourceAttributes searches for resources and passes the appropriate attributes to the walker function
func (r *Runner) WalkResourceAttributes(resource, attributeName string, walker func(*hcl.Attribute) error) error {
	for _, resource := range r.lookupResourcesByType(resource) {
		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name: attributeName,
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		if attribute, ok := body.Attributes[attributeName]; ok {
			log.Printf("[DEBUG] Walk `%s` attribute", resource.Type+"."+resource.Name+"."+attributeName)
			err := walker(attribute)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EnsureNoError is a helper for processing when no error occurs
// This function skips processing without returning an error to the caller when the error is warning
func (r *Runner) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if appErr, ok := err.(*Error); ok {
		switch appErr.Level {
		case WarningLevel:
			return nil
		case ErrorLevel:
			return appErr
		default:
			panic(appErr)
		}
	} else {
		return err
	}
}

// prepareVariableValues prepares Terraform variables from configs, input variables and environment variables.
// Variables in the configuration are overwritten by environment variables.
// Finally, they are overwritten by received input variable on the received order.
// Therefore, CLI flag input variables must be passed at the end of arguments.
// This is the responsibility of the caller.
// See https://www.terraform.io/intro/getting-started/variables.html#assigning-variables
func prepareVariableValues(configVars map[string]*configs.Variable, variables ...terraform.InputValues) map[string]map[string]cty.Value {
	overrideVariables := terraform.DefaultVariableValues(configVars).Override(getTFEnvVariables()).Override(variables...)

	variableValues := make(map[string]map[string]cty.Value)
	variableValues[""] = make(map[string]cty.Value)
	for k, iv := range overrideVariables {
		variableValues[""][k] = iv.Value
	}
	return variableValues
}

func (r *Runner) lookupResourcesByType(resourceType string) []*configs.Resource {
	ret := []*configs.Resource{}

	for _, resource := range r.TFConfig.Module.ManagedResources {
		if resource.Type == resourceType {
			ret = append(ret, resource)
		}
	}

	return ret
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
