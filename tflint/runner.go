package tflint

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/terraform/terraform"
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

	primaries             []*hcl.File
	overrides             []*hcl.File
	earlyDecodedResources map[string]*hclext.Block
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
func NewRunner(c *Config, files map[string]*hcl.File, ants map[string]Annotations, cfg *configs.Config, variables ...terraform.InputValues) (*Runner, error) {
	path := "root"
	if !cfg.Path.IsRoot() {
		path = cfg.Path.String()
	}
	log.Printf("[INFO] Initialize new runner for %s", path)

	variableValues, diags := prepareVariableValues(cfg, variables...)
	if diags.HasErrors() {
		return nil, diags
	}
	ctx := terraform.BuiltinEvalContext{
		Evaluator: &terraform.Evaluator{
			Meta: &terraform.ContextMeta{
				Env: getTFWorkspace(),
			},
			Config:             cfg.Root,
			VariableValues:     variableValues,
			VariableValuesLock: &sync.Mutex{},
		},
	}

	// Classify HCL files
	primaries := []*hcl.File{}
	overrides := []*hcl.File{}
	for name, f := range files {
		if filepath.Dir(name) != filepath.Clean(cfg.Module.SourceDir) {
			continue
		}

		if name == "override.tf" || name == "override.tf.json" || strings.HasSuffix(name, "_override.tf") || strings.HasSuffix(name, "_override.tf.json") {
			overrides = append(overrides, f)
		} else {
			primaries = append(primaries, f)
		}
	}
	sort.Slice(overrides, func(i, j int) bool {
		return overrides[i].Body.MissingItemRange().Filename < overrides[j].Body.MissingItemRange().Filename
	})

	runner := &Runner{
		TFConfig: cfg,
		Issues:   Issues{},

		// TODO: As described in the godoc for UnkeyedInstanceShim,
		// it will need to be replaced now that module.for_each is supported
		ctx:         ctx.WithPath(cfg.Path.UnkeyedInstanceShim()),
		files:       files,
		annotations: ants,
		config:      c,

		primaries:             primaries,
		overrides:             overrides,
		earlyDecodedResources: map[string]*hclext.Block{},
	}

	// Decode resource with count/for_each early
	bodyS := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "count"},
						{Name: "for_each"},
					},
				},
			},
		},
	}
	content := &hclext.BodyContent{}
	for _, f := range primaries {
		c, d := hclext.PartialContent(f.Body, bodyS)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = append(content.Blocks, c.Blocks...)
	}
	for _, f := range overrides {
		c, d := hclext.PartialContent(f.Body, bodyS)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = overrideBlocks(content.Blocks, c.Blocks)
	}
	for _, resource := range content.Blocks {
		ok, err := runner.willEvaluateResourceBlock(resource)
		if err != nil {
			return runner, err
		}
		if ok {
			runner.earlyDecodedResources[fmt.Sprintf("%s.%s", resource.Labels[0], resource.Labels[1])] = resource
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
		if parent.TFConfig.Path.IsRoot() && parent.config.IgnoreModules[moduleCall.SourceAddrRaw] {
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
			err := fmt.Errorf(
				"attribute of module not allowed was found in %s:%d; %w",
				moduleCall.DeclRange.Filename,
				moduleCall.DeclRange.Start.Line,
				causeErr,
			)
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
						err := fmt.Errorf(
							"failed to eval an expression in %s:%d; %w",
							attribute.Expr.Range().Filename,
							attribute.Expr.Range().Start.Line,
							diags.Err(),
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

func (r *Runner) GetModuleContent(bodyS *hclext.BodySchema) (*hclext.BodyContent, hcl.Diagnostics) {
	// TODO: early return if resource not found
	bodyS = appendDynamicBlockSchema(bodyS)
	content := &hclext.BodyContent{}
	diags := hcl.Diagnostics{}

	for _, f := range r.primaries {
		c, d := hclext.PartialContent(f.Body, bodyS)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = append(content.Blocks, c.Blocks...)
	}
	for _, f := range r.overrides {
		c, d := hclext.PartialContent(f.Body, bodyS)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = overrideBlocks(content.Blocks, c.Blocks)
	}

	content = resolveDynamicBlocks(content)

	out := &hclext.BodyContent{Attributes: content.Attributes}
	for _, block := range content.Blocks {
		if block.Type == "resource" {
			if _, exists := r.earlyDecodedResources[fmt.Sprintf("%s.%s", block.Labels[0], block.Labels[1])]; !exists {
				log.Printf("[WARN] Skip walking `%s` because it may not be created", block.Labels[0]+"."+block.Labels[1])
				continue
			}
		}

		out.Blocks = append(out.Blocks, block)
	}

	return out, diags
}

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

func resolveDynamicBlocks(content *hclext.BodyContent) *hclext.BodyContent {
	out := &hclext.BodyContent{Attributes: content.Attributes}

	for _, block := range content.Blocks {
		block.Body = resolveDynamicBlocks(block.Body)

		if block.Type != "dynamic" {
			out.Blocks = append(out.Blocks, block)
		} else {
			for _, dynamicContent := range block.Body.Blocks {
				block.Type = block.Labels[0]
				out.Blocks = append(out.Blocks, dynamicContent)
			}
		}
	}

	return out
}

func overrideBlocks(primaries, overrides hclext.Blocks) hclext.Blocks {
	dict := map[string]*hclext.Block{}
	for _, primary := range primaries {
		key := fmt.Sprintf("%s[%s]", primary.Type, strings.Join(primary.Labels, ","))
		dict[key] = primary
	}

	for _, override := range overrides {
		key := fmt.Sprintf("%s[%s]", override.Type, strings.Join(override.Labels, ","))
		if primary, exists := dict[key]; exists {
			for name, attr := range override.Body.Attributes {
				primary.Body.Attributes[name] = attr
			}

			for _, block := range override.Body.Blocks {
				primary.Body.Blocks = overrideBlocks(primary.Body.Blocks, block.Body.Blocks)
			}
		}
	}

	return primaries
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

	if errors.Is(err, sdk.ErrNullValue) || errors.Is(err, sdk.ErrUnevaluable) || errors.Is(err, sdk.ErrUnknownValue) {
		return nil
	}
	return err
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

func (r *Runner) ConfigFile() *hcl.File {
	return r.config.file
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
func prepareVariableValues(config *configs.Config, cliVars ...terraform.InputValues) (map[string]map[string]cty.Value, hcl.Diagnostics) {
	moduleKey := config.Path.UnkeyedInstanceShim().String()
	variableValues := make(map[string]map[string]cty.Value)
	variableValues[moduleKey] = make(map[string]cty.Value)

	configVars := map[string]*configs.Variable{}
	for k, v := range config.Module.Variables {
		configVars[k] = v
		// If default is not set, Terraform will interactively collect the variable values. Therefore, Evaluator returns the value as it is, even if default is not set.
		// This means that variables without default will be null in TFLint. This is unintended behavior, so assign an unknown value here.
		if v.Default == cty.NilVal {
			configVars[k].Default = cty.UnknownVal(v.Type)
		}
	}

	variables := terraform.DefaultVariableValues(configVars)
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
