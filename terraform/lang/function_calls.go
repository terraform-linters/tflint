package lang

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
)

// FunctionCall represents a function call in an HCL expression.
// The difference with hclsyntax.FunctionCallExpr is that
// function calls are also available in JSON syntax.
type FunctionCall struct {
	Name      string
	ArgsCount int
}

// FunctionCallsInExpr finds all of the function calls in the given expression.
func FunctionCallsInExpr(expr hcl.Expression) ([]*FunctionCall, hcl.Diagnostics) {
	if expr == nil {
		return nil, nil
	}

	// For JSON syntax, walker is not implemented,
	// so extract the hclsyntax.Node that we can walk on.
	// See https://github.com/hashicorp/hcl/issues/543
	nodes, diags := walkableNodesInExpr(expr)
	ret := []*FunctionCall{}

	for _, node := range nodes {
		visitDiags := hclsyntax.VisitAll(node, func(n hclsyntax.Node) hcl.Diagnostics {
			if funcCallExpr, ok := n.(*hclsyntax.FunctionCallExpr); ok {
				ret = append(ret, &FunctionCall{
					Name:      funcCallExpr.Name,
					ArgsCount: len(funcCallExpr.Args),
				})
			}
			return nil
		})
		diags = diags.Extend(visitDiags)
	}
	return ret, diags
}

// IsProviderDefined returns true if the function is provider-defined.
func (f *FunctionCall) IsProviderDefined() bool {
	return strings.HasPrefix(f.Name, "provider::")
}

// walkableNodesInExpr returns hclsyntax.Node from the given expression.
// If the expression is an hclsyntax expression, it is returned as is.
// If the expression is a JSON expression, it is parsed and
// hclsyntax.Node it contains is returned.
func walkableNodesInExpr(expr hcl.Expression) ([]hclsyntax.Node, hcl.Diagnostics) {
	nodes := []hclsyntax.Node{}

	expr = hcl.UnwrapExpressionUntil(expr, func(expr hcl.Expression) bool {
		_, native := expr.(hclsyntax.Expression)
		return native || json.IsJSONExpression(expr)
	})
	if expr == nil {
		return nil, nil
	}

	if json.IsJSONExpression(expr) {
		// HACK: For JSON expressions, we can get the JSON value as a literal
		//       without any prior HCL parsing by evaluating it in a nil context.
		//       We can take advantage of this property to walk through cty.Value
		//       that may contain HCL expressions instead of walking through
		//       expression nodes directly.
		//       See https://github.com/hashicorp/hcl/issues/642
		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return nodes, diags
		}

		err := cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
			if v.Type() != cty.String || v.IsNull() || !v.IsKnown() {
				return true, nil
			}

			node, parseDiags := hclsyntax.ParseTemplate([]byte(v.AsString()), expr.Range().Filename, expr.Range().Start)
			if diags.HasErrors() {
				diags = diags.Extend(parseDiags)
				return true, nil
			}

			nodes = append(nodes, node)
			return true, nil
		})
		if err != nil {
			return nodes, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Failed to walk the expression value",
				Detail:   err.Error(),
				Subject:  expr.Range().Ptr(),
			}}
		}

		return nodes, diags
	}

	// The JSON syntax is already processed, so it's guaranteed to be native syntax.
	nodes = append(nodes, expr.(hclsyntax.Expression))

	return nodes, nil
}
