package tflint

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/lang"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// Runner checks templates according rules.
// For variables interplation, it has Terraform eval context.
// After checking, it accumulates results as issues.
type Runner struct {
	TFConfig *configs.Config
	Issues   Issues

	ctx         terraform.EvalContext
	files       map[string]*hcl.File
	annotations map[string]Annotations
	config      *Config
	currentExpr hcl.Expression
	modVars     map[string]*moduleVariable
}

// Rule is interface for building the issue
type Rule interface {
	Name() string
	Severity() string
	Link() string
}

// NewRunner returns new TFLint runner
// It prepares built-in context (workpace metadata, variables) from
// received `configs.Config` and `terraform.InputValues`
func NewRunner(c *Config, files map[string]*hcl.File, ants map[string]Annotations, cfg *configs.Config, variables ...terraform.InputValues) (*Runner, error) {
	path := "root"
	if !cfg.Path.IsRoot() {
		path = cfg.Path.String()
	}
	log.Printf("[INFO] Initialize new runner for %s", path)

	ctx := terraform.BuiltinEvalContext{
		Evaluator: &terraform.Evaluator{
			Meta: &terraform.ContextMeta{
				Env: getTFWorkspace(),
			},
			Config:             cfg.Root,
			VariableValues:     prepareVariableValues(cfg, variables...),
			VariableValuesLock: &sync.Mutex{},
		},
	}

	runner := &Runner{
		TFConfig: cfg,
		Issues:   Issues{},

		// TODO: As described in the godoc for UnkeyedInstanceShim,
		// it will need to be replaced now that module.for_each is supported
		ctx:         ctx.WithPath(cfg.Path.UnkeyedInstanceShim()),
		files:       files,
		annotations: ants,
		config:      c,
	}

	return runner, nil
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
		if parent.TFConfig.Path.IsRoot() && parent.config.IgnoreModules[moduleCall.SourceAddr] {
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
					moduleCall.DeclRange.Filename,
					moduleCall.DeclRange.Start.Line,
				),
				Cause: causeErr,
			}
			log.Printf("[ERROR] %s", err)
			return runners, err
		}

		modVars := map[string]*moduleVariable{}
		for varName, rawVar := range cfg.Module.Variables {
			if attribute, exists := attributes[varName]; exists {
				evalauble, err := isEvaluableExpr(attribute.Expr)
				if err != nil {
					return runners, err
				}

				if evalauble {
					val, diags := parent.ctx.EvaluateExpr(attribute.Expr, cty.DynamicPseudoType, nil)
					if diags.HasErrors() {
						err := &Error{
							Code:  EvaluationError,
							Level: ErrorLevel,
							Message: fmt.Sprintf(
								"Failed to eval an expression in %s:%d",
								attribute.Expr.Range().Filename,
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

				if parent.TFConfig.Path.IsRoot() {
					modVars[varName] = &moduleVariable{
						Root:      true,
						DeclRange: attribute.Expr.Range(),
					}
				} else {
					parentVars := []*moduleVariable{}
					for _, ref := range listVarRefs(attribute.Expr) {
						if parentVar, exists := parent.modVars[ref.Name]; exists {
							parentVars = append(parentVars, parentVar)
						}
					}
					modVars[varName] = &moduleVariable{
						Parents:   parentVars,
						DeclRange: attribute.Expr.Range(),
					}
				}
			}
		}

		runner, err := NewRunner(parent.config, parent.files, parent.annotations, cfg)
		if err != nil {
			return runners, err
		}
		runner.modVars = modVars
		runners = append(runners, runner)
		moudleRunners, err := NewModuleRunners(runner)
		if err != nil {
			return runners, err
		}
		runners = append(runners, moudleRunners...)
	}

	return runners, nil
}

// TFConfigPath is a wrapper of addrs.Module
func (r *Runner) TFConfigPath() string {
	if r.TFConfig.Path.IsRoot() {
		return "root"
	}
	return r.TFConfig.Path.String()
}

// LookupIssues returns issues according to the received files
func (r *Runner) LookupIssues(files ...string) Issues {
	if len(files) == 0 {
		return r.Issues
	}

	issues := Issues{}
	for _, issue := range r.Issues {
		for _, file := range files {
			if file == issue.Range.Filename {
				issues = append(issues, issue)
			}
		}
	}
	return issues
}

// File returns the raw *hcl.File representation of a Terraform configuration at the specified path,
// or nil if there path does not match any configuration.
func (r *Runner) File(path string) *hcl.File {
	return r.files[path]
}

// Files returns the raw *hcl.File representation of all Terraform configuration in the module directory.
func (r *Runner) Files() map[string]*hcl.File {
	result := make(map[string]*hcl.File)
	for name, file := range r.files {
		if filepath.Dir(name) == filepath.Clean(r.TFConfig.Module.SourceDir) {
			result[name] = file
		}
	}
	return result
}

// Backend returns the backend configuration.
func (r *Runner) Backend() *configs.Backend {
	return r.TFConfig.Module.Backend
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

// IsNullExpr check the passed expression is null
func (r *Runner) IsNullExpr(expr hcl.Expression) (bool, error) {
	evaluable, err := isEvaluableExpr(expr)
	if err != nil {
		return false, err
	}

	if !evaluable {
		return false, nil
	}
	val, diags := r.ctx.EvaluateExpr(expr, cty.DynamicPseudoType, nil)
	if diags.HasErrors() {
		return false, diags.Err()
	}
	return val.IsNull(), nil
}

// LookupResourcesByType returns `configs.Resource` list according to the resource type
func (r *Runner) LookupResourcesByType(resourceType string) []*configs.Resource {
	ret := []*configs.Resource{}

	for _, resource := range r.TFConfig.Module.ManagedResources {
		if resource.Type == resourceType {
			ret = append(ret, resource)
		}
	}

	return ret
}

// EmitIssue builds an issue and accumulates it
func (r *Runner) EmitIssue(rule Rule, message string, location hcl.Range) {
	if r.TFConfig.Path.IsRoot() {
		r.emitIssue(&Issue{
			Rule:    rule,
			Message: message,
			Range:   location,
		})
	} else {
		for _, modVar := range r.listModuleVars(r.currentExpr) {
			r.emitIssue(&Issue{
				Rule:    rule,
				Message: message,
				Range:   modVar.DeclRange,
				Callers: append(modVar.callers(), location),
			})
		}
	}
}

// WithExpressionContext sets the context of the passed expression currently being processed.
func (r *Runner) WithExpressionContext(expr hcl.Expression, proc func() error) error {
	r.currentExpr = expr
	err := proc()
	r.currentExpr = nil
	return err
}

// DecodeRuleConfig extracts the rule's configuration into the given value
func (r *Runner) DecodeRuleConfig(ruleName string, val interface{}) error {
	if rule, exists := r.config.Rules[ruleName]; exists {
		diags := gohcl.DecodeBody(rule.Body, nil, val)
		if diags.HasErrors() {
			// HACK: If you enable the rule through the CLI instead of the file, its hcl.Body will not contain valid range.
			// @see https://github.com/hashicorp/hcl/blob/v2.5.0/merged.go#L132-L135
			if rule.Body.MissingItemRange().Filename == "<empty>" {
				return errors.New("This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration")
			}
			return diags
		}
	}
	return nil
}

// RuleConfig returns the corresponding rule configuration
func (r *Runner) RuleConfig(ruleName string) *RuleConfig {
	return r.config.Rules[ruleName]
}

func (r *Runner) emitIssue(issue *Issue) {
	if annotations, ok := r.annotations[issue.Range.Filename]; ok {
		for _, annotation := range annotations {
			if annotation.IsAffected(issue) {
				log.Printf("[INFO] %s (%s) is ignored by %s", issue.Range.String(), issue.Rule.Name(), annotation.String())
				return
			}
		}
	}
	r.Issues = append(r.Issues, issue)
}

func (r *Runner) listModuleVars(expr hcl.Expression) []*moduleVariable {
	ret := []*moduleVariable{}
	for _, ref := range listVarRefs(expr) {
		if modVar, exists := r.modVars[ref.Name]; exists {
			ret = append(ret, modVar.roots()...)
		}
	}
	return ret
}

// prepareVariableValues builds variableValues from configs, input variables and environment variables.
// Variables which declared in the configuration are overwritten by environment variables.
// Finally, they are overwritten by input variables in the order passed.
// Therefore, CLI flag input variables must be passed at the end of arguments.
// This is the responsibility of the caller.
// See https://learn.hashicorp.com/terraform/getting-started/variables.html#assigning-variables
func prepareVariableValues(config *configs.Config, variables ...terraform.InputValues) map[string]map[string]cty.Value {
	moduleKey := config.Path.UnkeyedInstanceShim().String()
	overrideVariables := terraform.DefaultVariableValues(config.Module.Variables).Override(getTFEnvVariables()).Override(variables...)

	variableValues := make(map[string]map[string]cty.Value)
	variableValues[moduleKey] = make(map[string]cty.Value)
	for k, iv := range overrideVariables {
		variableValues[moduleKey][k] = iv.Value
	}
	return variableValues
}

func listVarRefs(expr hcl.Expression) map[string]addrs.InputVariable {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		// Maybe this is bug
		panic(diags.Err())
	}

	ret := map[string]addrs.InputVariable{}
	for _, ref := range refs {
		if varRef, ok := ref.Subject.(addrs.InputVariable); ok {
			ret[varRef.String()] = varRef
		}
	}

	return ret
}
