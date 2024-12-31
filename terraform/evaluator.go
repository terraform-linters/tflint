package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agext/levenshtein"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type ContextMeta struct {
	Env                string
	OriginalWorkingDir string
}

type Evaluator struct {
	Meta           *ContextMeta
	ModulePath     addrs.ModuleInstance
	Config         *Config
	VariableValues map[string]map[string]cty.Value
}

// EvaluateExpr takes the given HCL expression and evaluates it to produce a value.
func (e *Evaluator) EvaluateExpr(expr hcl.Expression, wantType cty.Type) (cty.Value, hcl.Diagnostics) {
	if e == nil {
		panic("evaluator must not be nil")
	}
	return e.scope().EvalExpr(expr, wantType)
}

// ExpandBlock expands "dynamic" blocks and resources/modules with count/for_each.
//
// In the expanded body, the content can be retrieved with the HCL API without
// being aware of the differences in the dynamic block schema. Also, the number
// of blocks and attribute values will be the same as the expanded result.
func (e *Evaluator) ExpandBlock(body hcl.Body, schema *hclext.BodySchema) (hcl.Body, hcl.Diagnostics) {
	if e == nil {
		return body, nil
	}
	return e.scope().ExpandBlock(body, schema)
}

// scope creates a new evaluation scope.
// The difference with Evaluator is that each evaluation is independent
// and is not shared between goroutines.
func (e *Evaluator) scope() *lang.Scope {
	scope := &lang.Scope{CallStack: lang.NewCallStack()}
	scope.Data = &evaluationData{
		Scope:          scope,
		Meta:           e.Meta,
		ModulePath:     e.ModulePath,
		Config:         e.Config,
		VariableValues: e.VariableValues,
	}
	return scope
}

type evaluationData struct {
	Scope          *lang.Scope
	Meta           *ContextMeta
	ModulePath     addrs.ModuleInstance
	Config         *Config
	VariableValues map[string]map[string]cty.Value
}

var _ lang.Data = (*evaluationData)(nil)

func (d *evaluationData) GetCountAttr(addr addrs.CountAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	// Note that the actual evaluation of count.index is not done here.
	// count.index is already evaluated when expanded by ExpandBlock,
	// and the value is bound to the expanded body.
	//
	// Although, there are cases where count.index is evaluated as-is,
	// such as when not expanding the body. In that case, evaluate it
	// as an unknown and skip further checks.
	return cty.UnknownVal(cty.Number), nil
}

func (d *evaluationData) GetForEachAttr(addr addrs.ForEachAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	// Note that the actual evaluation of each.key/each.value is not done here.
	// each.key/each.value is already evaluated when expanded by ExpandBlock,
	// and the value is bound to the expanded body.
	//
	// Although, there are cases where each.key/each.value is evaluated as-is,
	// such as when not expanding the body. In that case, evaluate it
	// as an unknown and skip further checks.
	return cty.DynamicVal, nil
}

func (d *evaluationData) GetInputVariable(addr addrs.InputVariable, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	moduleConfig := d.Config.DescendentForInstance(d.ModulePath)
	if moduleConfig == nil {
		// should never happen, since we can't be evaluating in a module
		// that wasn't mentioned in configuration.
		panic(fmt.Sprintf("input variable read from %s, which has no configuration", d.ModulePath))
	}

	config := moduleConfig.Module.Variables[addr.Name]
	if config == nil {
		var suggestions []string
		for k := range moduleConfig.Module.Variables {
			suggestions = append(suggestions, k)
		}
		suggestion := nameSuggestion(addr.Name, suggestions)
		if suggestion != "" {
			suggestion = fmt.Sprintf(" Did you mean %q?", suggestion)
		} else {
			suggestion = fmt.Sprintf(" This variable can be declared with a variable %q {} block.", addr.Name)
		}

		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Reference to undeclared input variable`,
			Detail:   fmt.Sprintf(`An input variable with the name %q has not been declared.%s`, addr.Name, suggestion),
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags
	}

	moduleAddrStr := d.ModulePath.String()
	vals := d.VariableValues[moduleAddrStr]
	if vals == nil {
		return cty.UnknownVal(config.Type), diags
	}

	// In Terraform, it is the responsibility of the VariableTransformer
	// to convert the variable to the "final value", including the type conversion.
	// However, since TFLint does not preprocess variables by Graph Builder,
	// type conversion and default value are applied by Evaluator as in Terraform v1.1.
	//
	// This has some restrictions on the representation of dynamic variables compared
	// to Terraform, but since TFLint is intended for static analysis, this is often enough.
	val, isSet := vals[addr.Name]
	switch {
	case !isSet:
		// The config loader will ensure there is a default if the value is not
		// set at all.
		val = config.Default

	case val.IsNull() && !config.Nullable && config.Default != cty.NilVal:
		// If nullable=false a null value will use the configured default.
		val = config.Default
	}

	// Apply defaults from the variable's type constraint to the value,
	// unless the value is null. We do not apply defaults to top-level
	// null values, as doing so could prevent assigning null to a nullable
	// variable.
	if config.TypeDefaults != nil && !val.IsNull() {
		val = config.TypeDefaults.Apply(val)
	}

	var err error
	val, err = convert.Convert(val, config.ConstraintType)
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Incorrect variable type`,
			Detail:   fmt.Sprintf(`The resolved value of variable %q is not appropriate: %s.`, addr.Name, err),
			Subject:  &config.DeclRange,
		})
		val = cty.UnknownVal(config.Type)
	}

	if config.Sensitive {
		val = val.Mark(marks.Sensitive)
	}
	if config.Ephemeral {
		val = val.Mark(marks.Ephemeral)
	}

	return val, diags
}

func (d *evaluationData) GetLocalValue(addr addrs.LocalValue, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// First we'll make sure the requested value is declared in configuration,
	// so we can produce a nice message if not.
	moduleConfig := d.Config.DescendentForInstance(d.ModulePath)
	if moduleConfig == nil {
		// should never happen, since we can't be evaluating in a module
		// that wasn't mentioned in configuration.
		panic(fmt.Sprintf("local value read from %s, which has no configuration", d.ModulePath))
	}

	config := moduleConfig.Module.Locals[addr.Name]
	if config == nil {
		var suggestions []string
		for k := range moduleConfig.Module.Locals {
			suggestions = append(suggestions, k)
		}
		suggestion := nameSuggestion(addr.Name, suggestions)
		if suggestion != "" {
			suggestion = fmt.Sprintf(" Did you mean %q?", suggestion)
		}

		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Reference to undeclared local value`,
			Detail:   fmt.Sprintf(`A local value with the name %q has not been declared.%s`, addr.Name, suggestion),
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags
	}

	// Build a call stack for circular reference detection only when getting a local value.
	if diags := d.Scope.CallStack.Push(addrs.Reference{Subject: addr, SourceRange: rng}); diags.HasErrors() {
		return cty.UnknownVal(cty.DynamicPseudoType), diags
	}

	val, diags := d.Scope.EvalExpr(config.Expr, cty.DynamicPseudoType)

	d.Scope.CallStack.Pop()
	return val, diags
}

func (d *evaluationData) GetPathAttr(addr addrs.PathAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	switch addr.Name {

	case "cwd":
		var err error
		var wd string
		if d.Meta != nil {
			// Meta is always non-nil in the normal case, but some test cases
			// are not so realistic.
			wd = d.Meta.OriginalWorkingDir
		}
		if wd == "" {
			wd, err = os.Getwd()
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  `Failed to get working directory`,
					Detail:   fmt.Sprintf(`The value for path.cwd cannot be determined due to a system error: %s`, err),
					Subject:  rng.Ptr(),
				})
				return cty.DynamicVal, diags
			}
		}
		// The current working directory should always be absolute, whether we
		// just looked it up or whether we were relying on ContextMeta's
		// (possibly non-normalized) path.
		wd, err = filepath.Abs(wd)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Failed to get working directory`,
				Detail:   fmt.Sprintf(`The value for path.cwd cannot be determined due to a system error: %s`, err),
				Subject:  rng.Ptr(),
			})
			return cty.DynamicVal, diags
		}

		return cty.StringVal(filepath.ToSlash(wd)), diags

	case "module":
		moduleConfig := d.Config.DescendentForInstance(d.ModulePath)
		if moduleConfig == nil {
			// should never happen, since we can't be evaluating in a module
			// that wasn't mentioned in configuration.
			panic(fmt.Sprintf("module.path read from module %s, which has no configuration", d.ModulePath))
		}
		sourceDir := moduleConfig.Module.SourceDir
		return cty.StringVal(filepath.ToSlash(sourceDir)), diags

	case "root":
		sourceDir := d.Config.Module.SourceDir
		return cty.StringVal(filepath.ToSlash(sourceDir)), diags

	default:
		suggestion := nameSuggestion(addr.Name, []string{"cwd", "module", "root"})
		if suggestion != "" {
			suggestion = fmt.Sprintf(" Did you mean %q?", suggestion)
		}
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "path" attribute`,
			Detail:   fmt.Sprintf(`The "path" object does not have an attribute named %q.%s`, addr.Name, suggestion),
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags
	}
}

func (d *evaluationData) GetTerraformAttr(addr addrs.TerraformAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	switch addr.Name {

	case "workspace":
		workspaceName := d.Meta.Env
		return cty.StringVal(workspaceName), diags

	case "env":
		// Prior to Terraform 0.12 there was an attribute "env", which was
		// an alias name for "workspace". This was deprecated and is now
		// removed.
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "terraform" attribute`,
			Detail:   `The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The "state environment" concept was renamed to "workspace" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.`,
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags

	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "terraform" attribute`,
			Detail:   fmt.Sprintf(`The "terraform" object does not have an attribute named %q. The only supported attribute is terraform.workspace, the name of the currently-selected workspace.`, addr.Name),
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags
	}
}

// nameSuggestion tries to find a name from the given slice of suggested names
// that is close to the given name and returns it if found. If no suggestion
// is close enough, returns the empty string.
//
// The suggestions are tried in order, so earlier suggestions take precedence
// if the given string is similar to two or more suggestions.
//
// This function is intended to be used with a relatively-small number of
// suggestions. It's not optimized for hundreds or thousands of them.
func nameSuggestion(given string, suggestions []string) string {
	for _, suggestion := range suggestions {
		dist := levenshtein.Distance(given, suggestion, nil)
		if dist < 3 { // threshold determined experimentally
			return suggestion
		}
	}
	return ""
}
