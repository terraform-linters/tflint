package tflint

import (
	"fmt"
	"log"
	"path/filepath"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/zclconf/go-cty/cty"
)

// Runner checks templates according rules.
// For variables interplation, it has Terraform eval context.
// After checking, it accumulates results as issues.
type Runner struct {
	TFConfig *terraform.Config
	Issues   Issues
	Ctx      *terraform.Evaluator

	annotations map[string]Annotations
	config      *Config
	currentExpr hcl.Expression
	modVars     map[string]*moduleVariable
}

// Rule is interface for building the issue
type Rule interface {
	Name() string
	Severity() Severity
	Link() string
}

// NewRunner returns new TFLint runner
// It prepares built-in context (workpace metadata, variables) from
// received `configs.Config` and `terraform.InputValues`
func NewRunner(c *Config, ants map[string]Annotations, cfg *terraform.Config, variables ...terraform.InputValues) (*Runner, error) {
	path := "root"
	if !cfg.Path.IsRoot() {
		path = cfg.Path.String()
	}
	log.Printf("[INFO] Initialize new runner for %s", path)

	variableValues, diags := terraform.VariableValues(cfg, variables...)
	if diags.HasErrors() {
		return nil, diags
	}
	ctx := &terraform.Evaluator{
		Meta:           &terraform.ContextMeta{Env: terraform.Workspace()},
		ModulePath:     cfg.Path.UnkeyedInstanceShim(),
		Config:         cfg.Root,
		VariableValues: variableValues,
		CallGraph:      terraform.NewCallGraph(),
	}

	runner := &Runner{
		TFConfig: cfg,
		Issues:   Issues{},

		Ctx:         ctx,
		annotations: ants,
		config:      c,
	}

	return runner, nil
}

// NewModuleRunners returns new TFLint runners for child modules
// Recursively search modules and generate Runners
// In order to propagate attributes of moduleCall as variables to the module,
// evaluate the variables. If it cannot be evaluated, treat it as unknown
// Modules that are not evaluated (`count` is 0 or `for_each` is empty) are ignored.
func NewModuleRunners(parent *Runner) ([]*Runner, error) {
	runners := []*Runner{}

	for name, cfg := range parent.TFConfig.Children {
		moduleCall, ok := parent.TFConfig.Module.ModuleCalls[name]
		if !ok {
			panic(fmt.Errorf("Expected module call `%s` is not found in `%s`", name, parent.TFConfig.Path.String()))
		}
		if parent.TFConfig.Path.IsRoot() && parent.config.IgnoreModules[moduleCall.SourceAddrRaw] {
			log.Printf("[INFO] Ignore `%s` module", moduleCall.Name)
			continue
		}

		moduleCallSchema := &hclext.BodySchema{
			Blocks: []hclext.BlockSchema{
				{
					Type:       "module",
					LabelNames: []string{"name"},
					Body: &hclext.BodySchema{
						Attributes: []hclext.AttributeSchema{},
					},
				},
			},
		}
		for _, v := range cfg.Module.Variables {
			attr := hclext.AttributeSchema{Name: v.Name}
			moduleCallSchema.Blocks[0].Body.Attributes = append(moduleCallSchema.Blocks[0].Body.Attributes, attr)
		}

		moduleCalls, diags := parent.TFConfig.Module.PartialContent(moduleCallSchema, parent.Ctx)
		if diags.HasErrors() {
			return runners, diags
		}
		var moduleCallBodies []*hclext.BodyContent
		for _, block := range moduleCalls.Blocks {
			if moduleCall.Name == block.Labels[0] {
				moduleCallBodies = append(moduleCallBodies, block.Body)
			}
		}

		for _, body := range moduleCallBodies {
			modVars := map[string]*moduleVariable{}
			inputs := terraform.InputValues{}
			for varName, attribute := range body.Attributes {
				val, diags := parent.Ctx.EvaluateExpr(attribute.Expr, cty.DynamicPseudoType, terraform.EvalDataForNoInstanceKey)
				if diags.HasErrors() {
					err := fmt.Errorf(
						"failed to eval an expression in %s:%d; %w",
						attribute.Expr.Range().Filename,
						attribute.Expr.Range().Start.Line,
						diags,
					)
					log.Printf("[ERROR] %s", err)
					return runners, err
				}
				inputs[varName] = &terraform.InputValue{Value: val}

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

			runner, err := NewRunner(parent.config, parent.annotations, cfg, inputs)
			if err != nil {
				return runners, err
			}
			runner.modVars = modVars
			runners = append(runners, runner)
			moduleRunners, err := NewModuleRunners(runner)
			if err != nil {
				return runners, err
			}
			runners = append(runners, moduleRunners...)
		}
	}

	return runners, nil
}

// LookupIssues returns issues according to the received files
func (r *Runner) LookupIssues(files ...string) Issues {
	if len(files) == 0 {
		return r.Issues
	}

	issues := Issues{}
	for _, issue := range r.Issues {
		for _, file := range files {
			if filepath.Clean(file) == filepath.Clean(issue.Range.Filename) {
				issues = append(issues, issue)
			}
		}
	}
	return issues
}

// File returns the raw *hcl.File representation of a Terraform configuration at the specified path,
// or nil if there path does not match any configuration.
func (r *Runner) File(path string) *hcl.File {
	return r.TFConfig.Module.Files[path]
}

// Files returns the raw *hcl.File representation of all Terraform configuration in the module directory.
func (r *Runner) Files() map[string]*hcl.File {
	result := make(map[string]*hcl.File)
	for name, file := range r.TFConfig.Module.Files {
		if filepath.Dir(name) == filepath.Clean(r.TFConfig.Module.SourceDir) {
			result[name] = file
		}
	}
	return result
}

// Sources returns the sources in the module directory.
func (r *Runner) Sources() map[string][]byte {
	return r.TFConfig.Module.Sources
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

// RuleConfig returns the corresponding rule configuration
func (r *Runner) RuleConfig(ruleName string) *RuleConfig {
	return r.config.Rules[ruleName]
}

// ConfigSources returns the sources of TFLint config files
func (r *Runner) ConfigSources() map[string][]byte {
	return r.config.Sources()
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

func listVarRefs(expr hcl.Expression) map[string]addrs.InputVariable {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		// Maybe this is bug
		panic(diags)
	}

	ret := map[string]addrs.InputVariable{}
	for _, ref := range refs {
		if varRef, ok := ref.Subject.(addrs.InputVariable); ok {
			ret[varRef.String()] = varRef
		}
	}

	return ret
}
