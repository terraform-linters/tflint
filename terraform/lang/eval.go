package lang

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
	"github.com/terraform-linters/tflint/terraform/tfhcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// ExpandBlock expands "dynamic" blocks and resources/modules with count/for_each.
// Note that Terraform only expands dynamic blocks, but TFLint also expands
// count/for_each here.
//
// Expressions in expanded blocks are evaluated immediately, so all variables and
// function calls contained in attributes specified in the body schema are gathered.
func (s *Scope) ExpandBlock(body hcl.Body, schema *hclext.BodySchema) (hcl.Body, hcl.Diagnostics) {
	traversals := tfhcl.ExpandVariablesHCLExt(body, schema)
	refs, diags := References(traversals)

	exprs := tfhcl.ExpandExpressionsHCLExt(body, schema)
	funcCalls := []*FunctionCall{}
	for _, expr := range exprs {
		calls, funcDiags := FunctionCallsInExpr(expr)
		diags = diags.Extend(funcDiags)
		funcCalls = append(funcCalls, calls...)
	}

	ctx, ctxDiags := s.EvalContext(refs, funcCalls)
	diags = diags.Extend(ctxDiags)

	return tfhcl.Expand(body, ctx), diags
}

// EvalExpr evaluates a single expression in the receiving context and returns
// the resulting value. The value will be converted to the given type before
// it is returned if possible, or else an error diagnostic will be produced
// describing the conversion error.
//
// Pass an expected type of cty.DynamicPseudoType to skip automatic conversion
// and just obtain the returned value directly.
//
// If the returned diagnostics contains errors then the result may be
// incomplete, but will always be of the requested type.
func (s *Scope) EvalExpr(expr hcl.Expression, wantType cty.Type) (cty.Value, hcl.Diagnostics) {
	refs, diags := ReferencesInExpr(expr)
	funcCalls, funcDiags := FunctionCallsInExpr(expr)
	diags = diags.Extend(funcDiags)

	ctx, ctxDiags := s.EvalContext(refs, funcCalls)
	diags = diags.Extend(ctxDiags)
	if diags.HasErrors() {
		// We'll stop early if we found problems in the references, because
		// it's likely evaluation will produce redundant copies of the same errors.
		return cty.UnknownVal(wantType), diags
	}

	val, evalDiags := expr.Value(ctx)
	diags = diags.Extend(evalDiags)

	if wantType != cty.DynamicPseudoType {
		var convErr error
		val, convErr = convert.Convert(val, wantType)
		if convErr != nil {
			val = cty.UnknownVal(wantType)
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "Incorrect value type",
				Detail:      fmt.Sprintf("Invalid expression value: %s.", tfdiags.FormatError(convErr)),
				Subject:     expr.Range().Ptr(),
				Expression:  expr,
				EvalContext: ctx,
			})
		}
	}

	return val, diags
}

// EvalContext constructs a HCL expression evaluation context whose variable
// scope contains sufficient values to satisfy the given set of references
// and function calls.
//
// Most callers should prefer to use the evaluation helper methods that
// this type offers, but this is here for less common situations where the
// caller will handle the evaluation calls itself.
func (s *Scope) EvalContext(refs []*addrs.Reference, funcCalls []*FunctionCall) (*hcl.EvalContext, hcl.Diagnostics) {
	return s.evalContext(refs, s.SelfAddr, funcCalls)
}

func (s *Scope) evalContext(refs []*addrs.Reference, selfAddr addrs.Referenceable, funcCalls []*FunctionCall) (*hcl.EvalContext, hcl.Diagnostics) {
	if s == nil {
		panic("attempt to construct EvalContext for nil Scope")
	}

	var diags hcl.Diagnostics
	vals := make(map[string]cty.Value)
	funcs := s.Functions()
	// Provider-defined functions introduced in Terraform v1.8 cannot be
	// evaluated statically in many cases. Here, we avoid the error by dynamically
	// generating an evaluation context in which the provider-defined functions
	// in the given expression are replaced with mock functions.
	for _, call := range funcCalls {
		if !call.IsProviderDefined() {
			continue
		}
		// Some provider-defined functions are supported,
		// so only generate mocks for undefined functions
		if _, exists := funcs[call.Name]; !exists {
			funcs[call.Name] = NewMockFunction(call)
		}
	}
	ctx := &hcl.EvalContext{
		Variables: vals,
		Functions: funcs,
	}

	if len(refs) == 0 {
		// Easy path for common case where there are no references at all.
		return ctx, diags
	}

	// The reference set we are given has not been de-duped, and so there can
	// be redundant requests in it for two reasons:
	//  - The same item is referenced multiple times
	//  - Both an item and that item's container are separately referenced.
	// We will still visit every reference here and ask our data source for
	// it, since that allows us to gather a full set of any errors and
	// warnings, but once we've gathered all the data we'll then skip anything
	// that's redundant in the process of populating our values map.
	managedResources := map[string]cty.Value{}
	inputVariables := map[string]cty.Value{}
	localValues := map[string]cty.Value{}
	pathAttrs := map[string]cty.Value{}
	terraformAttrs := map[string]cty.Value{}
	countAttrs := map[string]cty.Value{}
	forEachAttrs := map[string]cty.Value{}

	for _, ref := range refs {
		rng := ref.SourceRange

		rawSubj := ref.Subject

		// This type switch must cover all of the "Referenceable" implementations
		// in package addrs, however we are removing the possibility of
		// Instances beforehand.
		switch addr := rawSubj.(type) {
		case addrs.ResourceInstance:
			rawSubj = addr.ContainingResource()
		}

		switch subj := rawSubj.(type) {
		case addrs.Resource:
			// Managed resources are not supported by TFLint, but it does support arbitrary
			// key names, so it gathers the referenced resource names.
			if subj.Mode != addrs.ManagedResourceMode {
				continue
			}
			managedResources[subj.Type] = cty.UnknownVal(cty.DynamicPseudoType)

		case addrs.InputVariable:
			val, valDiags := normalizeRefValue(s.Data.GetInputVariable(subj, rng))
			diags = diags.Extend(valDiags)
			inputVariables[subj.Name] = val

		case addrs.LocalValue:
			val, valDiags := normalizeRefValue(s.Data.GetLocalValue(subj, rng))
			diags = diags.Extend(valDiags)
			localValues[subj.Name] = val

		case addrs.PathAttr:
			val, valDiags := normalizeRefValue(s.Data.GetPathAttr(subj, rng))
			diags = diags.Extend(valDiags)
			pathAttrs[subj.Name] = val

		case addrs.TerraformAttr:
			val, valDiags := normalizeRefValue(s.Data.GetTerraformAttr(subj, rng))
			diags = diags.Extend(valDiags)
			terraformAttrs[subj.Name] = val

		case addrs.CountAttr:
			val, valDiags := normalizeRefValue(s.Data.GetCountAttr(subj, rng))
			diags = diags.Extend(valDiags)
			countAttrs[subj.Name] = val

		case addrs.ForEachAttr:
			val, valDiags := normalizeRefValue(s.Data.GetForEachAttr(subj, rng))
			diags = diags.Extend(valDiags)
			forEachAttrs[subj.Name] = val
		}
	}

	// Managed resources are exposed in two different locations. This is
	// at the top level where the resource type name is the root of the
	// traversal.
	for k, v := range managedResources {
		vals[k] = v
	}

	vals["var"] = cty.ObjectVal(inputVariables)
	vals["local"] = cty.ObjectVal(localValues)
	vals["path"] = cty.ObjectVal(pathAttrs)
	vals["terraform"] = cty.ObjectVal(terraformAttrs)
	vals["count"] = cty.ObjectVal(countAttrs)
	vals["each"] = cty.ObjectVal(forEachAttrs)

	// The following are unknown values as they are not supported by TFLint.
	vals["resource"] = cty.UnknownVal(cty.DynamicPseudoType)
	vals["data"] = cty.UnknownVal(cty.DynamicPseudoType)
	vals["module"] = cty.UnknownVal(cty.DynamicPseudoType)
	vals["self"] = cty.UnknownVal(cty.DynamicPseudoType)

	return ctx, diags
}

func normalizeRefValue(val cty.Value, diags hcl.Diagnostics) (cty.Value, hcl.Diagnostics) {
	if diags.HasErrors() {
		// If there are errors then we will force an unknown result so that
		// we can still evaluate and catch type errors but we'll avoid
		// producing redundant re-statements of the same errors we've already
		// dealt with here.
		return cty.UnknownVal(val.Type()), diags
	}
	return val, diags
}
