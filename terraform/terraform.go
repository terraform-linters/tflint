package terraform

import (
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

// ModuleCall represents a "module" block.
type ModuleCall struct {
	Name        string
	DefRange    hcl.Range
	Source      string
	SourceAttr  *hclext.Attribute
	Version     version.Constraints
	VersionAttr *hclext.Attribute
}

// @see https://github.com/hashicorp/terraform/blob/v1.2.7/internal/configs/module_call.go#L36-L224
func decodeModuleCall(block *hclext.Block) (*ModuleCall, hcl.Diagnostics) {
	module := &ModuleCall{
		Name:     block.Labels[0],
		DefRange: block.DefRange,
	}
	diags := hcl.Diagnostics{}

	if source, exists := block.Body.Attributes["source"]; exists {
		module.SourceAttr = source
		sourceDiags := gohcl.DecodeExpression(source.Expr, nil, &module.Source)
		diags = diags.Extend(sourceDiags)
	}

	if versionAttr, exists := block.Body.Attributes["version"]; exists {
		module.VersionAttr = versionAttr

		var versionVal string
		versionDiags := gohcl.DecodeExpression(versionAttr.Expr, nil, &versionVal)
		diags = diags.Extend(versionDiags)
		if diags.HasErrors() {
			return module, diags
		}

		constraints, err := version.NewConstraint(versionVal)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid version constraint",
				Detail:   "This string does not use correct version constraint syntax.",
				Subject:  versionAttr.Expr.Range().Ptr(),
			})
		}
		module.Version = constraints
	}

	return module, diags
}

// Local represents a single entry from a "locals" block.
type Local struct {
	Name     string
	DefRange hcl.Range
}

// ProviderRef represents a reference to a provider like `provider = google.europe` in a resource or module.
type ProviderRef struct {
	Name     string
	DefRange hcl.Range
}

// @see https://github.com/hashicorp/terraform/blob/v1.2.7/internal/configs/resource.go#L624-L695
func decodeProviderRef(expr hcl.Expression, defRange hcl.Range) (*ProviderRef, hcl.Diagnostics) {
	expr, diags := shimTraversalInString(expr)
	if diags.HasErrors() {
		return nil, diags
	}

	traversal, diags := hcl.AbsTraversalForExpr(expr)
	if diags.HasErrors() {
		return nil, diags
	}

	return &ProviderRef{
		Name:     traversal.RootName(),
		DefRange: defRange,
	}, nil
}

// @see https://github.com/hashicorp/terraform/blob/v1.2.5/internal/configs/compat_shim.go#L34
func shimTraversalInString(expr hcl.Expression) (hcl.Expression, hcl.Diagnostics) {
	// ObjectConsKeyExpr is a special wrapper type used for keys on object
	// constructors to deal with the fact that naked identifiers are normally
	// handled as "bareword" strings rather than as variable references. Since
	// we know we're interpreting as a traversal anyway (and thus it won't
	// matter whether it's a string or an identifier) we can safely just unwrap
	// here and then process whatever we find inside as normal.
	if ocke, ok := expr.(*hclsyntax.ObjectConsKeyExpr); ok {
		expr = ocke.Wrapped
	}

	if _, ok := expr.(*hclsyntax.TemplateExpr); !ok {
		return expr, nil
	}

	strVal, diags := expr.Value(nil)
	if diags.HasErrors() || strVal.IsNull() || !strVal.IsKnown() {
		// Since we're not even able to attempt a shim here, we'll discard
		// the diagnostics we saw so far and let the caller's own error
		// handling take care of reporting the invalid expression.
		return expr, nil
	}

	// The position handling here isn't _quite_ right because it won't
	// take into account any escape sequences in the literal string, but
	// it should be close enough for any error reporting to make sense.
	srcRange := expr.Range()
	startPos := srcRange.Start // copy
	startPos.Column++          // skip initial quote
	startPos.Byte++            // skip initial quote

	traversal, tDiags := hclsyntax.ParseTraversalAbs(
		[]byte(strVal.AsString()),
		srcRange.Filename,
		startPos,
	)
	diags = append(diags, tDiags...)

	return &hclsyntax.ScopeTraversalExpr{
		Traversal: traversal,
		SrcRange:  srcRange,
	}, diags
}
