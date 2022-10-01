package terraform

import (
	"fmt"
	"os"
	"path/filepath"

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

type Evaluator struct {
	Meta           *ContextMeta
	ModulePath     addrs.ModuleInstance
	Config         *Config
	VariableValues map[string]map[string]cty.Value
}

func (e *Evaluator) EvaluateExpr(expr hcl.Expression, wantType cty.Type) (cty.Value, hcl.Diagnostics) {
	scope := &lang.Scope{
		Data: &evaluationData{
			Evaluator:  e,
			ModulePath: e.ModulePath,
		},
	}
	return scope.EvalExpr(expr, wantType)
}

// ResourceIsEvaluable checks whether the passed resource meta-arguments
// (count/for_each) indicate the resource will be evaluated.
//
// If `count` is 0 or `for_each` is empty, Terraform will not evaluate
// the attributes of that resource. Terraform doesn't expect to pass null
// for these attributes (it will cause an error), but we'll treat them as
// if they were undefined.
func (e *Evaluator) ResourceIsEvaluable(resource *Resource) (bool, hcl.Diagnostics) {
	if resource.Count != nil {
		return e.countIsEvaluable(resource.Count)
	}

	if resource.ForEach != nil {
		return e.forEachIsEvaluable(resource.ForEach)
	}

	// If `count` or `for_each` is not defined, it will be evaluated by default
	return true, hcl.Diagnostics{}
}

func (e *Evaluator) ModuleCallIsEvaluable(moduleCall *ModuleCall) (bool, hcl.Diagnostics) {
	if moduleCall.Count != nil {
		return e.countIsEvaluable(moduleCall.Count)
	}

	if moduleCall.ForEach != nil {
		return e.forEachIsEvaluable(moduleCall.ForEach)
	}

	// If `count` or `for_each` is not defined, it will be evaluated by default
	return true, nil
}

func (e *Evaluator) countIsEvaluable(expr hcl.Expression) (bool, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	val, diags := e.EvaluateExpr(expr, cty.DynamicPseudoType)
	if diags.HasErrors() {
		return false, diags
	}
	val, _ = val.Unmark()

	if val.IsNull() {
		// null value means that attribute is not set
		return true, diags
	}
	if !val.IsKnown() {
		// unknown value is non-deterministic
		return false, diags
	}

	if val.Equals(cty.NumberIntVal(0)).True() {
		// `count = 0` is not evaluated
		return false, diags
	}
	// `count > 1` is evaluated`
	return true, diags
}

func (e *Evaluator) forEachIsEvaluable(expr hcl.Expression) (bool, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	val, diags := e.EvaluateExpr(expr, cty.DynamicPseudoType)
	if diags.HasErrors() {
		return false, diags
	}

	if val.IsNull() {
		// null value means that attribute is not set
		return true, diags
	}
	if !val.IsKnown() {
		// unknown value is non-deterministic
		return false, diags
	}
	if !val.CanIterateElements() {
		// uniteratable values (string, number, etc.) are
		return false, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "The `for_each` value is not iterable",
				Detail:   fmt.Sprintf("`%s` is not iterable", val.GoString()),
				Subject:  expr.Range().Ptr(),
			},
		}
	}
	if val.LengthInt() == 0 {
		// empty `for_each` is not evaluated
		return false, diags
	}
	// `for_each` with non-empty elements is evaluated
	return true, diags
}

type evaluationData struct {
	Evaluator  *Evaluator
	ModulePath addrs.ModuleInstance
}

var _ lang.Data = (*evaluationData)(nil)

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

	// Mark if sensitive
	if config.Sensitive {
		val = val.Mark(marks.Sensitive)
	}

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
