package tfhcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

type expandDynamicSpec struct {
	blockType      string
	blockTypeRange hcl.Range
	defRange       hcl.Range
	forEachVal     cty.Value
	iteratorName   string
	labelExprs     []hcl.Expression
	contentBody    hcl.Body
}

func (b *expandBody) decodeDynamicSpec(blockS *hcl.BlockHeaderSchema, rawSpec *hcl.Block) (*expandDynamicSpec, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	var schema *hcl.BodySchema
	if len(blockS.LabelNames) != 0 {
		schema = dynamicBlockBodySchemaLabels
	} else {
		schema = dynamicBlockBodySchemaNoLabels
	}

	specContent, specDiags := rawSpec.Body.Content(schema)
	diags = append(diags, specDiags...)
	if specDiags.HasErrors() {
		return nil, diags
	}

	//// for_each attribute

	eachAttr := specContent.Attributes["for_each"]
	eachVal, eachDiags := eachAttr.Expr.Value(b.ctx)
	diags = append(diags, eachDiags...)

	if !eachVal.CanIterateElements() && eachVal.Type() != cty.DynamicPseudoType {
		// We skip this error for DynamicPseudoType because that means we either
		// have a null (which is checked immediately below) or an unknown
		// (which is handled in the expandBody Content methods).
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Invalid dynamic for_each value",
			Detail:      fmt.Sprintf("Cannot use a %s value in for_each. An iterable collection is required.", eachVal.Type().FriendlyName()),
			Subject:     eachAttr.Expr.Range().Ptr(),
			Expression:  eachAttr.Expr,
			EvalContext: b.ctx,
		})
		return nil, diags
	}
	if eachVal.IsNull() {
		diags = append(diags, &hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "Invalid dynamic for_each value",
			Detail:      "Cannot use a null value in for_each.",
			Subject:     eachAttr.Expr.Range().Ptr(),
			Expression:  eachAttr.Expr,
			EvalContext: b.ctx,
		})
		return nil, diags
	}

	//// iterator attribute

	iteratorName := blockS.Type
	if iteratorAttr := specContent.Attributes["iterator"]; iteratorAttr != nil {
		itTraversal, itDiags := hcl.AbsTraversalForExpr(iteratorAttr.Expr)
		diags = append(diags, itDiags...)
		if itDiags.HasErrors() {
			return nil, diags
		}

		if len(itTraversal) != 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid dynamic iterator name",
				Detail:   "Dynamic iterator must be a single variable name.",
				Subject:  itTraversal.SourceRange().Ptr(),
			})
			return nil, diags
		}

		iteratorName = itTraversal.RootName()
	}

	var labelExprs []hcl.Expression
	if labelsAttr := specContent.Attributes["labels"]; labelsAttr != nil {
		var labelDiags hcl.Diagnostics
		labelExprs, labelDiags = hcl.ExprList(labelsAttr.Expr)
		diags = append(diags, labelDiags...)
		if labelDiags.HasErrors() {
			return nil, diags
		}

		if len(labelExprs) > len(blockS.LabelNames) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Extraneous dynamic block label",
				Detail:   fmt.Sprintf("Blocks of type %q require %d label(s).", blockS.Type, len(blockS.LabelNames)),
				Subject:  labelExprs[len(blockS.LabelNames)].Range().Ptr(),
			})
			return nil, diags
		} else if len(labelExprs) < len(blockS.LabelNames) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Insufficient dynamic block labels",
				Detail:   fmt.Sprintf("Blocks of type %q require %d label(s).", blockS.Type, len(blockS.LabelNames)),
				Subject:  labelsAttr.Expr.Range().Ptr(),
			})
			return nil, diags
		}
	}

	// Since our schema requests only blocks of type "content", we can assume
	// that all entries in specContent.Blocks are content blocks.
	if len(specContent.Blocks) == 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing dynamic content block",
			Detail:   "A dynamic block must have a nested block of type \"content\" to describe the body of each generated block.",
			Subject:  &specContent.MissingItemRange,
		})
		return nil, diags
	}
	if len(specContent.Blocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Extraneous dynamic content block",
			Detail:   "Only one nested content block is allowed for each dynamic block.",
			Subject:  &specContent.Blocks[1].DefRange,
		})
		return nil, diags
	}

	return &expandDynamicSpec{
		blockType:      blockS.Type,
		blockTypeRange: rawSpec.LabelRanges[0],
		defRange:       rawSpec.DefRange,
		forEachVal:     eachVal,
		iteratorName:   iteratorName,
		labelExprs:     labelExprs,
		contentBody:    specContent.Blocks[0].Body,
	}, diags
}

func (s *expandDynamicSpec) newBlock(i *dynamicIteration, ctx *hcl.EvalContext) (*hcl.Block, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var labels []string
	var labelRanges []hcl.Range
	lCtx := i.EvalContext(ctx)
	for _, labelExpr := range s.labelExprs {
		labelVal, labelDiags := labelExpr.Value(lCtx)
		diags = append(diags, labelDiags...)
		if labelDiags.HasErrors() {
			return nil, diags
		}

		var convErr error
		labelVal, convErr = convert.Convert(labelVal, cty.String)
		if convErr != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "Invalid dynamic block label",
				Detail:      fmt.Sprintf("Cannot use this value as a dynamic block label: %s.", convErr),
				Subject:     labelExpr.Range().Ptr(),
				Expression:  labelExpr,
				EvalContext: lCtx,
			})
			return nil, diags
		}
		if labelVal.IsNull() {
			diags = append(diags, &hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "Invalid dynamic block label",
				Detail:      "Cannot use a null value as a dynamic block label.",
				Subject:     labelExpr.Range().Ptr(),
				Expression:  labelExpr,
				EvalContext: lCtx,
			})
			return nil, diags
		}
		if !labelVal.IsKnown() {
			return nil, diags
		}
		if labelVal.IsMarked() {
			diags = append(diags, &hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "Invalid dynamic block label",
				Detail:      "Cannot use a marked value as a dynamic block label.",
				Subject:     labelExpr.Range().Ptr(),
				Expression:  labelExpr,
				EvalContext: lCtx,
			})
			return nil, diags
		}

		labels = append(labels, labelVal.AsString())
		labelRanges = append(labelRanges, labelExpr.Range())
	}

	block := &hcl.Block{
		Type:        s.blockType,
		TypeRange:   s.blockTypeRange,
		Labels:      labels,
		LabelRanges: labelRanges,
		DefRange:    s.defRange,
		Body:        s.contentBody,
	}

	return block, diags
}

type expandMetaArgSpec struct {
	rawBlock   *hcl.Block
	countSet   bool
	countVal   cty.Value
	countNum   int
	forEachSet bool
	forEachVal cty.Value
}

func (b *expandBody) decodeMetaArgSpec(rawSpec *hcl.Block) (*expandMetaArgSpec, hcl.Diagnostics) {
	spec := &expandMetaArgSpec{rawBlock: rawSpec}
	var diags hcl.Diagnostics

	specContent, _, specDiags := rawSpec.Body.PartialContent(expandableBlockBodySchema)
	diags = append(diags, specDiags...)
	if specDiags.HasErrors() {
		return spec, diags
	}

	//// count attribute

	if countAttr, exists := specContent.Attributes["count"]; exists {
		spec.countSet = true

		countVal, countDiags := countAttr.Expr.Value(b.ctx)
		diags = append(diags, countDiags...)
		countVal, _ = countVal.Unmark()

		spec.countVal = countVal

		// We skip validation for count attribute if the value is unknwon
		if countVal.IsKnown() {
			if countVal.IsNull() {
				diags = append(diags, &hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Invalid count argument",
					Detail:      `The given "count" argument value is null. An integer is required.`,
					Subject:     countAttr.Expr.Range().Ptr(),
					Expression:  countAttr.Expr,
					EvalContext: b.ctx,
				})
				return spec, diags
			}

			var convErr error
			countVal, convErr = convert.Convert(countVal, cty.Number)
			if convErr != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Incorrect value type",
					Detail:      fmt.Sprintf("Invalid expression value: %s.", tfdiags.FormatError(convErr)),
					Subject:     countAttr.Expr.Range().Ptr(),
					Expression:  countAttr.Expr,
					EvalContext: b.ctx,
				})
				return spec, diags
			}

			err := gocty.FromCtyValue(countVal, &spec.countNum)
			if err != nil {
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Invalid count argument",
					Detail:      fmt.Sprintf(`The given "count" argument value is unsuitable: %s.`, err),
					Subject:     countAttr.Expr.Range().Ptr(),
					Expression:  countAttr.Expr,
					EvalContext: b.ctx,
				})
				return spec, diags
			}
			if spec.countNum < 0 {
				diags = diags.Append(&hcl.Diagnostic{
					Severity:    hcl.DiagError,
					Summary:     "Invalid count argument",
					Detail:      `The given "count" argument value is unsuitable: negative numbers are not supported.`,
					Subject:     countAttr.Expr.Range().Ptr(),
					Expression:  countAttr.Expr,
					EvalContext: b.ctx,
				})
				return spec, diags
			}
		}
	}

	//// for_each attribute

	if eachAttr, exists := specContent.Attributes["for_each"]; exists {
		spec.forEachSet = true

		eachVal, eachDiags := eachAttr.Expr.Value(b.ctx)
		diags = append(diags, eachDiags...)

		spec.forEachVal = eachVal

		if !eachVal.CanIterateElements() && eachVal.Type() != cty.DynamicPseudoType {
			// We skip this error for DynamicPseudoType because that means we either
			// have a null (which is checked immediately below) or an unknown
			// (which is handled in the expandBody Content methods).
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "The `for_each` value is not iterable",
				Detail:      fmt.Sprintf("`%s` is not iterable", eachVal.GoString()),
				Subject:     eachAttr.Expr.Range().Ptr(),
				Expression:  eachAttr.Expr,
				EvalContext: b.ctx,
			})
			return spec, diags
		}
		if eachVal.IsNull() {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "Invalid for_each argument",
				Detail:      `The given "for_each" argument value is unsuitable: the given "for_each" argument value is null. A map, or set of strings is allowed.`,
				Subject:     eachAttr.Expr.Range().Ptr(),
				Expression:  eachAttr.Expr,
				EvalContext: b.ctx,
			})
			return spec, diags
		}
	}

	return spec, diags
}
