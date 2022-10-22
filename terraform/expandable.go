package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type expandable struct {
	Count   hcl.Expression
	ForEach hcl.Expression
}

// expandBlock returns multiple blocks based on the meta-arguments (count/for_each).
//
// This function returns no blocks if `count` is 0, if `for_each` is empty, or if they are unknown.
// Otherwise it returns the number of blocks according to the value.
//
// Expressions containing `count.*` or `each.*` are evaluated here when expanding blocks.
// Make an instance key and bind the evaluation result based on it to the expression.
// Note that sensitive values are not bound. This is a limitation in value decoding.
// This means that `count.*`, `each.*` with sensitive values will resolve to unknown values.
func (e *expandable) expandBlock(ctx *Evaluator, block *hclext.Block) (hclext.Blocks, hcl.Diagnostics) {
	if e.Count != nil {
		return e.expandBlockByCount(ctx, block)
	}

	if e.ForEach != nil {
		return e.expandBlockByForEach(ctx, block)
	}

	return hclext.Blocks{block}, hcl.Diagnostics{}
}

func (e *expandable) expandBlockByCount(ctx *Evaluator, block *hclext.Block) (hclext.Blocks, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	countVal, countDiags := ctx.EvaluateExpr(e.Count, cty.Number, EvalDataForNoInstanceKey)
	diags = diags.Extend(countDiags)
	if diags.HasErrors() {
		return hclext.Blocks{}, diags
	}
	countVal, _ = countVal.Unmark()

	if countVal.IsNull() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid count argument",
			Detail:   `The given "count" argument value is null. An integer is required.`,
			Subject:  e.Count.Range().Ptr(),
		})
		return hclext.Blocks{}, diags
	}
	if !countVal.IsKnown() {
		// If count is unknown, no blocks are returned
		return hclext.Blocks{}, diags
	}

	var count int
	err := gocty.FromCtyValue(countVal, &count)
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid count argument",
			Detail:   fmt.Sprintf(`The given "count" argument value is unsuitable: %s.`, err),
			Subject:  e.Count.Range().Ptr(),
		})
		return hclext.Blocks{}, diags
	}
	if count < 0 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid count argument",
			Detail:   `The given "count" argument value is unsuitable: negative numbers are not supported.`,
			Subject:  e.Count.Range().Ptr(),
		})
		return hclext.Blocks{}, diags
	}

	blocks := make(hclext.Blocks, count)
	for i := 0; i < count; i++ {
		expanded := block.Copy()
		keyData := InstanceKeyEvalData{CountIndex: cty.NumberIntVal(int64(i))}

		walkDiags := expanded.Body.WalkAttributes(func(attr *hclext.Attribute) hcl.Diagnostics {
			var diags hcl.Diagnostics

			refs, refsDiags := lang.ReferencesInExpr(attr.Expr)
			if refsDiags.HasErrors() {
				diags = diags.Extend(refsDiags)
				return diags
			}

			var contain bool
			for _, ref := range refs {
				if _, ok := ref.Subject.(addrs.CountAttr); ok {
					contain = true
				}
			}

			if contain {
				val, evalDiags := ctx.EvaluateExpr(attr.Expr, cty.DynamicPseudoType, keyData)
				if evalDiags.HasErrors() {
					diags = diags.Extend(evalDiags)
					return diags
				}
				// If marked as sensitive, the cty.Value cannot be marshaled in MessagePack,
				// so only bind it if it is unmarked.
				if !val.IsMarked() {
					// Even if there is no instance key later, the evaluated result is bound to
					// the expression so that it can be referenced by EvaluateExpr.
					attr.Expr = hclext.BindValue(val, attr.Expr)
				}
			}
			return diags
		})

		diags = diags.Extend(walkDiags)
		blocks[i] = expanded
	}

	return blocks, diags
}

func (e *expandable) expandBlockByForEach(ctx *Evaluator, block *hclext.Block) (hclext.Blocks, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	forEach, forEachDiags := ctx.EvaluateExpr(e.ForEach, cty.DynamicPseudoType, EvalDataForNoInstanceKey)
	diags = diags.Extend(forEachDiags)
	if diags.HasErrors() {
		return hclext.Blocks{}, diags
	}

	if forEach.IsNull() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid for_each argument",
			Detail:   `The given "for_each" argument value is unsuitable: the given "for_each" argument value is null. A map, or set of strings is allowed.`,
			Subject:  e.ForEach.Range().Ptr(),
		})
		return hclext.Blocks{}, diags
	}
	if !forEach.IsKnown() {
		// If for_each is unknown, no blocks are returned
		return hclext.Blocks{}, diags
	}
	if !forEach.CanIterateElements() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "The `for_each` value is not iterable",
			Detail:   fmt.Sprintf("`%s` is not iterable", forEach.GoString()),
			Subject:  e.ForEach.Range().Ptr(),
		})
		return hclext.Blocks{}, diags
	}

	blocks := make(hclext.Blocks, forEach.LengthInt())
	it := forEach.ElementIterator()
	for i := 0; it.Next(); i++ {
		expanded := block.Copy()

		key, value := it.Element()
		keyData := InstanceKeyEvalData{EachKey: key, EachValue: value}

		walkDiags := expanded.Body.WalkAttributes(func(attr *hclext.Attribute) hcl.Diagnostics {
			var diags hcl.Diagnostics

			refs, refsDiags := lang.ReferencesInExpr(attr.Expr)
			if refsDiags.HasErrors() {
				diags = diags.Extend(refsDiags)
				return diags
			}

			var contain bool
			for _, ref := range refs {
				if _, ok := ref.Subject.(addrs.ForEachAttr); ok {
					contain = true
				}
			}

			if contain {
				val, evalDiags := ctx.EvaluateExpr(attr.Expr, cty.DynamicPseudoType, keyData)
				if evalDiags.HasErrors() {
					diags = diags.Extend(evalDiags)
					return diags
				}
				// If marked as sensitive, the cty.Value cannot be marshaled in MessagePack,
				// so only bind it if it is unmarked.
				if !val.IsMarked() {
					// Even if there is no instance key later, the evaluated result is bound to
					// the expression so that it can be referenced by EvaluateExpr.
					attr.Expr = hclext.BindValue(val, attr.Expr)
				}
			}
			return diags
		})

		diags = diags.Extend(walkDiags)
		blocks[i] = expanded
	}

	return blocks, diags
}
