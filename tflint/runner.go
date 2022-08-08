package tflint

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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

	ctx         *terraform.Evaluator
	annotations map[string]Annotations
	config      *Config
	currentExpr hcl.Expression
	modVars     map[string]*moduleVariable

	earlyDecodedResources map[string]map[string]*hclext.Block
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

	variableValues, diags := prepareVariableValues(cfg, variables...)
	if diags.HasErrors() {
		return nil, diags
	}
	ctx := &terraform.Evaluator{
		Meta:           &terraform.ContextMeta{Env: getTFWorkspace()},
		ModulePath:     cfg.Path.UnkeyedInstanceShim(),
		Config:         cfg.Root,
		VariableValues: variableValues,
	}

	runner := &Runner{
		TFConfig: cfg,
		Issues:   Issues{},

		ctx:         ctx,
		annotations: ants,
		config:      c,

		earlyDecodedResources: map[string]map[string]*hclext.Block{},
	}

	for _, resource := range runner.TFConfig.Module.Resources {
		evaluable, err := runner.isEvaluableResource(resource)
		if err != nil {
			return runner, err
		}
		if evaluable {
			resourceType := resource.Labels[0]
			resourceName := resource.Labels[1]

			if _, exists := runner.earlyDecodedResources[resourceType]; !exists {
				runner.earlyDecodedResources[resourceType] = map[string]*hclext.Block{}
			}
			runner.earlyDecodedResources[resourceType][resourceName] = resource
		}
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
		evaluable, err := parent.isEvaluableModuleCall(moduleCall)
		if err != nil {
			return runners, fmt.Errorf(
				"failed to eval count/for_each meta-arguments in %s:%d; %w",
				moduleCall.DeclRange.Filename,
				moduleCall.DeclRange.Start.Line,
				err,
			)
		}
		if !evaluable {
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

		moduleCalls, diags := parent.TFConfig.Module.PartialContent(moduleCallSchema)
		if diags.HasErrors() {
			return runners, diags
		}
		var moduleCallBody *hclext.BodyContent
		for _, block := range moduleCalls.Blocks {
			if moduleCall.Name == block.Labels[0] {
				moduleCallBody = block.Body
			}
		}

		modVars := map[string]*moduleVariable{}
		for varName, attribute := range moduleCallBody.Attributes {
			if rawVar, exists := cfg.Module.Variables[varName]; exists {
				evalauble, err := isEvaluableExpr(attribute.Expr)
				if err != nil {
					return runners, err
				}

				if evalauble {
					val, diags := parent.ctx.EvaluateExpr(attribute.Expr, cty.DynamicPseudoType)
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
		runners = append(runners, runner)
		moudleRunners, err := NewModuleRunners(runner)
		if err != nil {
			return runners, err
		}
		runners = append(runners, moudleRunners...)
	}

	return runners, nil
}

// GetModuleContent extracts body content from Terraform configurations based on the passed schema.
// Basically, this function is a wrapper for hclext.PartialContent, but in some ways it reproduces
// Terraform language semantics.
//
//  1. The block schema implicitly adds dynamic blocks to the target
//     https://www.terraform.io/language/expressions/dynamic-blocks
//  2. Supports overriding files
//     https://www.terraform.io/language/files/override
//  3. Resources not created by count or for_each will be ignored
//     https://www.terraform.io/language/meta-arguments/count
//     https://www.terraform.io/language/meta-arguments/for_each
//
// However, this behavior is controlled by options. The above is the default.
func (r *Runner) GetModuleContent(bodyS *hclext.BodySchema, opts sdk.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
	// For performance, determine in advance whether the target resource exists.
	if opts.Hint.ResourceType != "" {
		if _, exists := r.earlyDecodedResources[opts.Hint.ResourceType]; !exists {
			return &hclext.BodyContent{}, nil
		}
	}

	bodyS = appendDynamicBlockSchema(bodyS)

	content, diags := r.TFConfig.Module.PartialContent(bodyS)
	if diags.HasErrors() {
		return content, diags
	}

	content = resolveDynamicBlocks(content)

	if opts.IncludeNotCreated {
		return content, diags
	}

	out := &hclext.BodyContent{Attributes: content.Attributes}
	for _, block := range content.Blocks {
		if block.Type == "resource" {
			resourceType := block.Labels[0]
			resourceName := block.Labels[1]

			if _, exists := r.earlyDecodedResources[resourceType]; !exists {
				log.Printf("[WARN] Skip walking `%s` because it may not be created", resourceType+"."+resourceName)
				continue
			}
			if _, exists := r.earlyDecodedResources[resourceType][resourceName]; !exists {
				log.Printf("[WARN] Skip walking `%s` because it may not be created", resourceType+"."+resourceName)
				continue
			}
		}

		out.Blocks = append(out.Blocks, block)
	}

	return out, diags
}

// appendDynamicBlockSchema appends a dynamic block schema to block schemes recursively.
// The content retrieved by the added schema is formatted by resolveDynamicBlocks in the same way as regular blocks.
func appendDynamicBlockSchema(schema *hclext.BodySchema) *hclext.BodySchema {
	out := &hclext.BodySchema{Attributes: schema.Attributes}

	for _, block := range schema.Blocks {
		block.Body = appendDynamicBlockSchema(block.Body)

		out.Blocks = append(out.Blocks, block, hclext.BlockSchema{
			Type:       "dynamic",
			LabelNames: []string{"name"},
			Body: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "content",
						Body: block.Body,
					},
				},
			},
		})
	}

	return out
}

// resolveDynamicBlocks formats the passed content based on the block schema added by appendDynamicBlockSchema.
// This allows you to get all named blocks without being aware of the difference in the structure of the dynamic block.
func resolveDynamicBlocks(content *hclext.BodyContent) *hclext.BodyContent {
	out := &hclext.BodyContent{Attributes: content.Attributes}

	for _, block := range content.Blocks {
		block.Body = resolveDynamicBlocks(block.Body)

		if block.Type != "dynamic" {
			out.Blocks = append(out.Blocks, block)
		} else {
			for _, dynamicContent := range block.Body.Blocks {
				dynamicContent.Type = block.Labels[0]
				out.Blocks = append(out.Blocks, dynamicContent)
			}
		}
	}

	return out
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

// DecodeRuleConfig extracts the rule's configuration into the given value
func (r *Runner) DecodeRuleConfig(ruleName string, val interface{}) error {
	if rule, exists := r.config.Rules[ruleName]; exists {
		// If you enable the rule through the CLI instead of the file, its hcl.Body will be nil.
		if rule.Body == nil {
			diags := gohcl.DecodeBody(hcl.EmptyBody(), nil, val)
			if diags.HasErrors() {
				return errors.New("This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration")
			}
		} else {
			diags := gohcl.DecodeBody(rule.Body, nil, val)
			if diags.HasErrors() {
				return diags
			}
		}
	}
	return nil
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

// prepareVariableValues builds variableValues from configs, input variables and environment variables.
// Variables which declared in the configuration are overwritten by environment variables.
// Finally, they are overwritten by input variables in the order passed.
// Therefore, CLI flag input variables must be passed at the end of arguments.
// This is the responsibility of the caller.
// See https://learn.hashicorp.com/terraform/getting-started/variables.html#assigning-variables
func prepareVariableValues(config *terraform.Config, cliVars ...terraform.InputValues) (map[string]map[string]cty.Value, hcl.Diagnostics) {
	moduleKey := config.Path.UnkeyedInstanceShim().String()
	variableValues := make(map[string]map[string]cty.Value)
	variableValues[moduleKey] = make(map[string]cty.Value)

	configVars := map[string]*terraform.Variable{}
	for k, v := range config.Module.Variables {
		configVars[k] = v
		// If default is not set, Terraform will interactively collect the variable values. Therefore, Evaluator returns the value as it is, even if default is not set.
		// This means that variables without default will be null in TFLint. This is unintended behavior, so assign an unknown value here.
		if v.Default == cty.NilVal {
			configVars[k].Default = cty.UnknownVal(v.Type)
		}
	}

	variables := DefaultVariableValues(configVars)
	envVars, diags := getTFEnvVariables(configVars)
	if diags.HasErrors() {
		return variableValues, diags
	}
	overrideVariables := variables.Override(envVars).Override(cliVars...)

	for k, iv := range overrideVariables {
		variableValues[moduleKey][k] = iv.Value
	}
	return variableValues, nil
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
