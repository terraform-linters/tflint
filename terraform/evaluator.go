package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agext/levenshtein"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type ContextMeta struct {
	Env                string
	OriginalWorkingDir string
}

type CallStack struct {
	addrs map[string]addrs.Reference
	stack []string
}

func NewCallStack() *CallStack {
	return &CallStack{
		addrs: make(map[string]addrs.Reference),
		stack: make([]string, 0),
	}
}

func (g *CallStack) Push(addr addrs.Reference) hcl.Diagnostics {
	g.stack = append(g.stack, addr.Subject.String())

	if _, exists := g.addrs[addr.Subject.String()]; exists {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "circular reference found",
				Detail:   g.String(),
				Subject:  addr.SourceRange.Ptr(),
			},
		}
	}
	g.addrs[addr.Subject.String()] = addr
	return hcl.Diagnostics{}
}

func (g *CallStack) Pop() {
	if g.Empty() {
		panic("cannot pop from empty stack")
	}

	addr := g.stack[len(g.stack)-1]
	g.stack = g.stack[:len(g.stack)-1]
	delete(g.addrs, addr)
}

func (g *CallStack) String() string {
	return strings.Join(g.stack, " -> ")
}

func (g *CallStack) Empty() bool {
	return len(g.stack) == 0
}

func (g *CallStack) Clear() {
	g.addrs = make(map[string]addrs.Reference)
	g.stack = make([]string, 0)
}

type Evaluator struct {
	Meta           *ContextMeta
	ModulePath     addrs.ModuleInstance
	Config         *Config
	VariableValues map[string]map[string]cty.Value
	CallStack      *CallStack
}

func (e *Evaluator) EvaluateExpr(expr hcl.Expression, wantType cty.Type, keyData InstanceKeyEvalData) (cty.Value, hcl.Diagnostics) {
	scope := &lang.Scope{
		Data: &evaluationData{
			Evaluator:       e,
			ModulePath:      e.ModulePath,
			InstanceKeyData: keyData,
		},
	}
	return scope.EvalExpr(expr, wantType)
}

type InstanceKeyEvalData struct {
	CountIndex         cty.Value
	EachKey, EachValue cty.Value
}

var EvalDataForNoInstanceKey = InstanceKeyEvalData{}

type evaluationData struct {
	Evaluator       *Evaluator
	ModulePath      addrs.ModuleInstance
	InstanceKeyData InstanceKeyEvalData
}

var _ lang.Data = (*evaluationData)(nil)

func (d *evaluationData) GetCountAttr(addr addrs.CountAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// Even when evaluating an expression that already has the value of `count.*` bound to it,
	// it still tries to create an EvalContext because it contains `count.*` as a reference.
	// In that case it returns an unknown value without returning an error.
	if d.InstanceKeyData == EvalDataForNoInstanceKey {
		return cty.UnknownVal(cty.Number), diags
	}

	switch addr.Name {

	case "index":
		idxVal := d.InstanceKeyData.CountIndex
		if idxVal == cty.NilVal {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Reference to "count" in non-counted context`,
				Detail:   `The "count" object can only be used in "module", "resource", and "data" blocks, and only when the "count" argument is set.`,
				Subject:  rng.Ptr(),
			})
			return cty.UnknownVal(cty.Number), diags
		}
		return idxVal, diags

	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "count" attribute`,
			Detail:   fmt.Sprintf(`The "count" object does not have an attribute named %q. The only supported attribute is count.index, which is the index of each instance of a resource block that has the "count" argument set.`, addr.Name),
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags
	}
}

func (d *evaluationData) GetForEachAttr(addr addrs.ForEachAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// Even when evaluating an expression that already has the value of `each.*` bound to it,
	// it still tries to create an EvalContext because it contains `each.*` as a reference.
	// In that case it returns an unknown value without returning an error.
	if d.InstanceKeyData == EvalDataForNoInstanceKey {
		return cty.UnknownVal(cty.DynamicPseudoType), diags
	}

	var returnVal cty.Value
	switch addr.Name {

	case "key":
		returnVal = d.InstanceKeyData.EachKey
	case "value":
		returnVal = d.InstanceKeyData.EachValue

		if returnVal == cty.NilVal {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `each.value cannot be used in this context`,
				Detail:   `A reference to "each.value" has been used in a context in which it unavailable, such as when the configuration no longer contains the value in its "for_each" expression. Remove this reference to each.value in your configuration to work around this error.`,
				Subject:  rng.Ptr(),
			})
			return cty.UnknownVal(cty.DynamicPseudoType), diags
		}
	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "each" attribute`,
			Detail:   fmt.Sprintf(`The "each" object does not have an attribute named %q. The supported attributes are each.key and each.value, the current key and value pair of the "for_each" attribute set.`, addr.Name),
			Subject:  rng.Ptr(),
		})
		return cty.DynamicVal, diags
	}

	if returnVal == cty.NilVal {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Reference to "each" in context without for_each`,
			Detail:   `The "each" object can be used only in "module" or "resource" blocks, and only when the "for_each" argument is set.`,
			Subject:  rng.Ptr(),
		})
		return cty.UnknownVal(cty.DynamicPseudoType), diags
	}
	return returnVal, diags
}

func (d *evaluationData) GetInputVariable(addr addrs.InputVariable, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	moduleConfig := d.Evaluator.Config.DescendentForInstance(d.ModulePath)
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
	vals := d.Evaluator.VariableValues[moduleAddrStr]
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

	// Apply defaults from the variable's type constraint to the value,
	// unless the value is null. We do not apply defaults to top-level
	// null values, as doing so could prevent assigning null to a nullable
	// variable.
	if config.TypeDefaults != nil && !val.IsNull() {
		val = config.TypeDefaults.Apply(val)
	}

	// Mark if sensitive
	if config.Sensitive {
		val = val.Mark(marks.Sensitive)
	}

	return val, diags
}

func (d *evaluationData) GetLocalValue(addr addrs.LocalValue, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// First we'll make sure the requested value is declared in configuration,
	// so we can produce a nice message if not.
	moduleConfig := d.Evaluator.Config.DescendentForInstance(d.ModulePath)
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
	if diags := d.Evaluator.CallStack.Push(addrs.Reference{Subject: addr, SourceRange: rng}); diags.HasErrors() {
		return cty.UnknownVal(cty.DynamicPseudoType), diags
	}

	// Always use EvalDataForNoInstanceKey because local values cannot use expressions
	// that depend on instance keys, such as `count.*` and `each.*`.
	val, diags := d.Evaluator.EvaluateExpr(config.Expr, cty.DynamicPseudoType, EvalDataForNoInstanceKey)

	d.Evaluator.CallStack.Pop()
	return val, diags
}

func (d *evaluationData) GetPathAttr(addr addrs.PathAttr, rng hcl.Range) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	switch addr.Name {

	case "cwd":
		var err error
		var wd string
		if d.Evaluator.Meta != nil {
			// Meta is always non-nil in the normal case, but some test cases
			// are not so realistic.
			wd = d.Evaluator.Meta.OriginalWorkingDir
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
		moduleConfig := d.Evaluator.Config.DescendentForInstance(d.ModulePath)
		if moduleConfig == nil {
			// should never happen, since we can't be evaluating in a module
			// that wasn't mentioned in configuration.
			panic(fmt.Sprintf("module.path read from module %s, which has no configuration", d.ModulePath))
		}
		sourceDir := moduleConfig.Module.SourceDir
		return cty.StringVal(filepath.ToSlash(sourceDir)), diags

	case "root":
		sourceDir := d.Evaluator.Config.Module.SourceDir
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
		workspaceName := d.Evaluator.Meta.Env
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
