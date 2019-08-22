package tflint

import (
	"errors"
	"fmt"
	"log"
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

	ctx         terraform.BuiltinEvalContext
	annotations map[string]Annotations
	config      *Config
	currentExpr hcl.Expression
	modVars     map[string]*moduleVariable
}

// Rule is interface for building the issue
type Rule interface {
	Name() string
	Type() string
	Link() string
}

// NewRunner returns new TFLint runner
// It prepares built-in context (workpace metadata, variables) from
// received `configs.Config` and `terraform.InputValues`
func NewRunner(c *Config, ants map[string]Annotations, cfg *configs.Config, variables ...terraform.InputValues) (*Runner, error) {
	path := "root"
	if !cfg.Path.IsRoot() {
		path = cfg.Path.String()
	}
	log.Printf("[INFO] Initialize new runner for %s", path)

	runner := &Runner{
		TFConfig:  cfg,
		Issues:    []*issue.Issue{},
		AwsClient: &client.AwsClient{},

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
		annotations: ants,
		config:      c,
	}

	// Initialize client for the root runner
	if c.DeepCheck && cfg.Path.IsRoot() {
		// FIXME: Alias providers are not considered
		providerConfig, err := NewProviderConfig(
			cfg.Module.ProviderConfigs["aws"],
			runner,
			client.AwsProviderBlockSchema,
		)
		if err != nil {
			return nil, err
		}
		creds, err := client.ConvertToCredentials(providerConfig)
		if err != nil {
			return nil, err
		}

		runner.AwsClient, err = client.NewAwsClient(c.AwsCredentials.Merge(creds))
		if err != nil {
			return nil, err
		}
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
				if isEvaluable(attribute.Expr) {
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

		runner, err := NewRunner(parent.config, parent.annotations, cfg)
		if err != nil {
			return runners, err
		}
		runner.modVars = modVars
		// Inherit parent's AwsClient
		runner.AwsClient = parent.AwsClient
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
				expr.Range().Filename,
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
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: diags.Err(),
		}
		log.Printf("[ERROR] %s", err)
		return err
	}

	err := cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			err := &Error{
				Code:  UnknownValueError,
				Level: WarningLevel,
				Message: fmt.Sprintf(
					"Unknown value found in %s:%d; Please use environment variables or tfvars to set the value",
					expr.Range().Filename,
					expr.Range().Start.Line,
				),
			}
			log.Printf("[WARN] %s; TFLint ignores an expression includes an unknown value.", err)
			return false, err
		}

		if v.IsNull() {
			err := &Error{
				Code:  NullValueError,
				Level: WarningLevel,
				Message: fmt.Sprintf(
					"Null value found in %s:%d",
					expr.Range().Filename,
					expr.Range().Start.Line,
				),
			}
			log.Printf("[WARN] %s; TFLint ignores an expression includes an null value.", err)
			return false, err
		}

		return true, nil
	})

	if err != nil {
		return err
	}

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
				expr.Range().Filename,
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
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return err
	}
	return nil
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
	for _, resource := range r.LookupResourcesByType(resource) {
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
			r.currentExpr = attribute.Expr
			err := walker(attribute)
			r.currentExpr = nil
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WalkResourceBlocks walks all blocks of the passed resource and invokes the passed function
func (r *Runner) WalkResourceBlocks(resource, blockType string, walker func(*hcl.Block) error) error {
	for _, resource := range r.LookupResourcesByType(resource) {
		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: blockType,
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, block := range body.Blocks {
			log.Printf("[DEBUG] Walk `%s` block", resource.Type+"."+resource.Name+"."+blockType)
			err := walker(block)
			if err != nil {
				return err
			}
		}

		// Walk in the same way for dynamic blocks. Note that we are not expanding blocks.
		// Therefore, expressions that use iterator are unevaluable.
		dynBody, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "dynamic",
					LabelNames: []string{"name"},
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, block := range dynBody.Blocks {
			if len(block.Labels) == 1 && block.Labels[0] == blockType {
				body, _, diags = block.Body.PartialContent(&hcl.BodySchema{
					Blocks: []hcl.BlockHeaderSchema{
						{
							Type: "content",
						},
					},
				})
				if diags.HasErrors() {
					return diags
				}

				for _, block := range body.Blocks {
					log.Printf("[DEBUG] Walk dynamic `%s` block", resource.Type+"."+resource.Name+"."+blockType)
					err := walker(block)
					if err != nil {
						return err
					}
				}
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

// IsNullExpr check the passed expression is null
func (r *Runner) IsNullExpr(expr hcl.Expression) bool {
	if !isEvaluable(expr) {
		return false
	}
	val, diags := r.ctx.EvaluateExpr(expr, cty.DynamicPseudoType, nil)
	if diags.HasErrors() {
		return false
	}
	return val.IsNull()
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

// EachStringSliceExprs iterates an evaluated value and the corresponding expression
// If the given expression is a static list, get an expression for each value
// If not, the given expression is used as it is
func (r *Runner) EachStringSliceExprs(expr hcl.Expression, proc func(val string, expr hcl.Expression)) error {
	var vals []string
	err := r.EvaluateExpr(expr, &vals)

	exprs, diags := hcl.ExprList(expr)
	if diags.HasErrors() {
		log.Printf("[DEBUG] Expr is not static list: %s", diags)
		for range vals {
			exprs = append(exprs, expr)
		}
	}

	return r.EnsureNoError(err, func() error {
		for idx, val := range vals {
			proc(val, exprs[idx])
		}
		return nil
	})
}

// EmitIssue builds an issue and accumulates it
func (r *Runner) EmitIssue(rule Rule, message string, location hcl.Range) {
	if r.TFConfig.Path.IsRoot() {
		r.emitIssue(&issue.Issue{
			Detector: rule.Name(),
			Type:     rule.Type(),
			Message:  message,
			Line:     location.Start.Line,
			File:     location.Filename,
			Link:     rule.Link(),
		})
	} else {
		for _, modVar := range r.listModuleVars(r.currentExpr) {
			r.emitIssue(&issue.Issue{
				Detector: rule.Name(),
				Type:     rule.Type(),
				Message:  message,
				Line:     modVar.DeclRange.Start.Line,
				File:     modVar.DeclRange.Filename,
				Link:     rule.Link(),
			})
		}
	}
}

func (r *Runner) emitIssue(issue *issue.Issue) {
	if annotations, ok := r.annotations[issue.File]; ok {
		for _, annotation := range annotations {
			if annotation.IsAffected(issue) {
				log.Printf("[INFO] %s:%d (%s) is ignored by %s", issue.File, issue.Line, issue.Detector, annotation.String())
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

func listVarRefs(expr hcl.Expression) []addrs.InputVariable {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		// Maybe this is bug
		panic(diags.Err())
	}

	ret := []addrs.InputVariable{}
	for _, ref := range refs {
		if varRef, ok := ref.Subject.(addrs.InputVariable); ok {
			ret = append(ret, varRef)
		}
	}

	return ret
}
