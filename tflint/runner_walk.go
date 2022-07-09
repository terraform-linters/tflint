package tflint

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// WalkExpressions visits all expressions, including those in the file before merging.
// Note that it behaves differently in native HCL syntax and JSON syntax.
// In the HCL syntax, expressions in expressions, such as list and object are passed to
// the walker function. The walker should check the type of the expression.
// In the JSON syntax, only an expression of an attribute seen from the top level of the file
// is passed, not expressions in expressions to the walker. This is an API limitation of JSON syntax.
func (r *Runner) WalkExpressions(walker func(hcl.Expression) error) error {
	visit := func(node hclsyntax.Node) hcl.Diagnostics {
		if expr, ok := node.(hcl.Expression); ok {
			if err := walker(expr); err != nil {
				// FIXME: walker should returns hcl.Diagnostics directly
				return hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  err.Error(),
					},
				}
			}
		}
		return hcl.Diagnostics{}
	}

	for _, file := range r.Files() {
		if body, ok := file.Body.(*hclsyntax.Body); ok {
			diags := hclsyntax.VisitAll(body, visit)
			if diags.HasErrors() {
				return diags
			}
			continue
		}

		// In JSON syntax, everything can be walked as an attribute.
		attrs, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			return diags
		}

		for _, attr := range attrs {
			if err := walker(attr.Expr); err != nil {
				return err
			}
		}
	}

	return nil
}
