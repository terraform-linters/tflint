package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/agext/levenshtein"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/instances"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/terraform/lang/marks"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
)

// Evaluator provides the necessary contextual data for evaluating expressions
// for a particular walk operation.
type Evaluator struct {
	// Meta is contextual metadata about the current operation.
	Meta *ContextMeta

	// Config is the root node in the configuration tree.
	Config *configs.Config

	// VariableValues is a map from variable names to their associated values,
	// within the module indicated by ModulePath. VariableValues is modified
	// concurrently, and so it must be accessed only while holding
	// VariableValuesLock.
	//
	// The first map level is string representations of addr.ModuleInstance
	// values, while the second level is variable names.
	VariableValues     map[string]map[string]cty.Value
	VariableValuesLock *sync.Mutex
}

// Scope creates an evaluation scope for the given module path and optional
// resource.
//
// If the "self" argument is nil then the "self" object is not available
// in evaluated expressions. Otherwise, it behaves as an alias for the given
// address.
func (e *Evaluator) Scope(data lang.Data, self addrs.Referenceable) *lang.Scope {
	return &lang.Scope{
		Data:     data,
		SelfAddr: self,
		PureOnly: true,
		BaseDir:  ".", // Always current working directory for now.
	}
}

// evaluationStateData is an implementation of lang.Data that resolves
// references primarily (but not exclusively) using information from a State.
type evaluationStateData struct {
	Evaluator *Evaluator

	// ModulePath is the path through the dynamic module tree to the module
	// that references will be resolved relative to.
	ModulePath addrs.ModuleInstance

	// InstanceKeyData describes the values, if any, that are accessible due
	// to repetition of a containing object using "count" or "for_each"
	// arguments. (It is _not_ used for the for_each inside "dynamic" blocks,
	// since the user specifies in that case which variable name to locally
	// shadow.)
	InstanceKeyData InstanceKeyEvalData
}

// InstanceKeyEvalData is the old name for instances.RepetitionData, aliased
// here for compatibility. In new code, use instances.RepetitionData instead.
type InstanceKeyEvalData = instances.RepetitionData

// EvalDataForNoInstanceKey is a value of InstanceKeyData that sets no instance
// key values at all, suitable for use in contexts where no keyed instance
// is relevant.
var EvalDataForNoInstanceKey = InstanceKeyEvalData{}

// evaluationStateData must implement lang.Data
var _ lang.Data = (*evaluationStateData)(nil)

func (d *evaluationStateData) GetInputVariable(addr addrs.InputVariable, rng tfdiags.SourceRange) (cty.Value, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics

	// First we'll make sure the requested value is declared in configuration,
	// so we can produce a nice message if not.
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
			Subject:  rng.ToHCL().Ptr(),
		})
		return cty.DynamicVal, diags
	}
	d.Evaluator.VariableValuesLock.Lock()
	defer d.Evaluator.VariableValuesLock.Unlock()

	moduleAddrStr := d.ModulePath.String()
	vals := d.Evaluator.VariableValues[moduleAddrStr]
	if vals == nil {
		return cty.UnknownVal(config.Type), diags
	}

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
		// We should never get here because this problem should've been caught
		// during earlier validation, but we'll do something reasonable anyway.
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

func (d *evaluationStateData) GetPathAttr(addr addrs.PathAttr, rng tfdiags.SourceRange) (cty.Value, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
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
					Subject:  rng.ToHCL().Ptr(),
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
				Subject:  rng.ToHCL().Ptr(),
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
			Subject:  rng.ToHCL().Ptr(),
		})
		return cty.DynamicVal, diags
	}
}

func (d *evaluationStateData) GetTerraformAttr(addr addrs.TerraformAttr, rng tfdiags.SourceRange) (cty.Value, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
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
			Subject:  rng.ToHCL().Ptr(),
		})
		return cty.DynamicVal, diags

	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Invalid "terraform" attribute`,
			Detail:   fmt.Sprintf(`The "terraform" object does not have an attribute named %q. The only supported attribute is terraform.workspace, the name of the currently-selected workspace.`, addr.Name),
			Subject:  rng.ToHCL().Ptr(),
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
