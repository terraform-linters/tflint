package tfhcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/zclconf/go-cty/cty"
)

// expandBody wraps another hcl.Body and expands any "dynamic" blocks, count/for-each
// resources found inside whenever Content or PartialContent is called.
type expandBody struct {
	original         hcl.Body
	ctx              *hcl.EvalContext
	dynamicIteration *dynamicIteration // non-nil if we're nested inside a "dynamic" block
	metaArgIteration *metaArgIteration // non-nil if we're nested inside a block with meta-arguments

	// These are used with PartialContent to produce a "remaining items"
	// body to return. They are nil on all bodies fresh out of the transformer.
	//
	// Note that this is re-implemented here rather than delegating to the
	// existing support required by the underlying body because we need to
	// retain access to the entire original body on subsequent decode operations
	// so we can retain any "dynamic" blocks for types we didn't take consume
	// on the first pass.
	hiddenAttrs  map[string]struct{}
	hiddenBlocks map[string]hcl.BlockHeaderSchema
}

func (b *expandBody) Content(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Diagnostics) {
	extSchema := b.extendSchema(schema)
	rawContent, diags := b.original.Content(extSchema)

	blocks, blockDiags := b.expandBlocks(schema, rawContent.Blocks, false)
	diags = append(diags, blockDiags...)
	attrs, attrDiags := b.prepareAttributes(rawContent.Attributes)
	diags = append(diags, attrDiags...)

	content := &hcl.BodyContent{
		Attributes:       attrs,
		Blocks:           blocks,
		MissingItemRange: b.original.MissingItemRange(),
	}

	return content, diags
}

func (b *expandBody) PartialContent(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Body, hcl.Diagnostics) {
	extSchema := b.extendSchema(schema)
	rawContent, _, diags := b.original.PartialContent(extSchema)
	// We discard the "remain" argument above because we're going to construct
	// our own remain that also takes into account remaining "dynamic" blocks.

	blocks, blockDiags := b.expandBlocks(schema, rawContent.Blocks, true)
	diags = append(diags, blockDiags...)
	attrs, attrDiags := b.prepareAttributes(rawContent.Attributes)
	diags = append(diags, attrDiags...)

	content := &hcl.BodyContent{
		Attributes:       attrs,
		Blocks:           blocks,
		MissingItemRange: b.original.MissingItemRange(),
	}

	remain := &expandBody{
		original:         b.original,
		ctx:              b.ctx,
		dynamicIteration: b.dynamicIteration,
		metaArgIteration: b.metaArgIteration,
		hiddenAttrs:      make(map[string]struct{}),
		hiddenBlocks:     make(map[string]hcl.BlockHeaderSchema),
	}
	for name := range b.hiddenAttrs {
		remain.hiddenAttrs[name] = struct{}{}
	}
	for typeName, blockS := range b.hiddenBlocks {
		remain.hiddenBlocks[typeName] = blockS
	}
	for _, attrS := range schema.Attributes {
		remain.hiddenAttrs[attrS.Name] = struct{}{}
	}
	for _, blockS := range schema.Blocks {
		remain.hiddenBlocks[blockS.Type] = blockS
	}

	return content, remain, diags
}

func (b *expandBody) extendSchema(schema *hcl.BodySchema) *hcl.BodySchema {
	// We augment the requested schema to also include our special "dynamic"
	// block type, since then we'll get instances of it interleaved with
	// all of the literal child blocks we must also include.
	extSchema := &hcl.BodySchema{
		Attributes: schema.Attributes,
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks), len(schema.Blocks)+len(b.hiddenBlocks)+1),
	}
	copy(extSchema.Blocks, schema.Blocks)
	extSchema.Blocks = append(extSchema.Blocks, dynamicBlockHeaderSchema)

	// If we have any hiddenBlocks then we also need to register those here
	// so that a call to "Content" on the underlying body won't fail.
	// (We'll filter these out again once we process the result of either
	// Content or PartialContent.)
	for _, blockS := range b.hiddenBlocks {
		extSchema.Blocks = append(extSchema.Blocks, blockS)
	}

	// If we have any hiddenAttrs then we also need to register these, for
	// the same reason as we deal with hiddenBlocks above.
	if len(b.hiddenAttrs) != 0 {
		newAttrs := make([]hcl.AttributeSchema, len(schema.Attributes), len(schema.Attributes)+len(b.hiddenAttrs))
		copy(newAttrs, extSchema.Attributes)
		for name := range b.hiddenAttrs {
			newAttrs = append(newAttrs, hcl.AttributeSchema{
				Name:     name,
				Required: false,
			})
		}
		extSchema.Attributes = newAttrs
	}

	return extSchema
}

func (b *expandBody) prepareAttributes(rawAttrs hcl.Attributes) (hcl.Attributes, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if len(b.hiddenAttrs) == 0 && b.dynamicIteration == nil && b.metaArgIteration == nil {
		// Easy path: just pass through the attrs from the original body verbatim
		return rawAttrs, diags
	}

	// Otherwise we have some work to do: we must filter out any attributes
	// that are hidden (since a previous PartialContent call already saw these)
	// and wrap the expressions of the inner attributes so that they will
	// have access to our iteration variables.
	attrs := make(hcl.Attributes, len(rawAttrs))
	for name, rawAttr := range rawAttrs {
		if _, hidden := b.hiddenAttrs[name]; hidden {
			continue
		}
		if b.dynamicIteration != nil || b.metaArgIteration != nil {
			attr := *rawAttr // shallow copy so we can mutate it
			expr := exprWrap{
				Expression: attr.Expr,
				di:         b.dynamicIteration,
				mi:         b.metaArgIteration,
			}
			// Unlike hcl/ext/dynblock, wrapped expressions are evaluated immediately.
			// The result is bound to the expression and can be accessed without
			// the iterator context.
			val, evalDiags := expr.Value(b.ctx)
			if evalDiags.HasErrors() {
				diags = append(diags, evalDiags...)
				continue
			}
			// Marked values (e.g. sensitive values) are unbound for serialization.
			if !val.ContainsMarked() {
				attr.Expr = hclext.BindValue(val, expr)
			}
			attrs[name] = &attr
		} else {
			// If we have no active iteration then no wrapping is required.
			attrs[name] = rawAttr
		}
	}
	return attrs, diags
}

func (b *expandBody) expandBlocks(schema *hcl.BodySchema, rawBlocks hcl.Blocks, partial bool) (hcl.Blocks, hcl.Diagnostics) {
	var blocks hcl.Blocks
	var diags hcl.Diagnostics

	for _, rawBlock := range rawBlocks {
		switch rawBlock.Type {
		case "dynamic":
			expandedBlocks, expandDiags := b.expandDynamicBlock(schema, rawBlock, partial)
			blocks = append(blocks, expandedBlocks...)
			diags = append(diags, expandDiags...)

		case "resource", "module":
			expandedBlocks, expandDiags := b.expandMetaArgBlock(schema, rawBlock)
			blocks = append(blocks, expandedBlocks...)
			diags = append(diags, expandDiags...)

		default:
			if _, hidden := b.hiddenBlocks[rawBlock.Type]; !hidden {
				blocks = append(blocks, b.expandStaticBlock(rawBlock))
			}
		}
	}

	return blocks, diags
}

func (b *expandBody) expandDynamicBlock(schema *hcl.BodySchema, rawBlock *hcl.Block, partial bool) (hcl.Blocks, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	realBlockType := rawBlock.Labels[0]
	if _, hidden := b.hiddenBlocks[realBlockType]; hidden {
		return hcl.Blocks{}, diags
	}

	var blockS *hcl.BlockHeaderSchema
	for _, candidate := range schema.Blocks {
		if candidate.Type == realBlockType {
			blockS = &candidate
			break
		}
	}
	if blockS == nil {
		// Not a block type that the caller requested.
		if !partial {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported block type",
				Detail:   fmt.Sprintf("Blocks of type %q are not expected here.", realBlockType),
				Subject:  &rawBlock.LabelRanges[0],
			})
		}
		return hcl.Blocks{}, diags
	}

	spec, specDiags := b.decodeDynamicSpec(blockS, rawBlock)
	diags = append(diags, specDiags...)
	if specDiags.HasErrors() {
		return hcl.Blocks{}, diags
	}

	if !spec.forEachVal.IsKnown() {
		// If for_each is unknown, no blocks are returned
		return hcl.Blocks{}, diags
	}

	var blocks hcl.Blocks

	for it := spec.forEachVal.ElementIterator(); it.Next(); {
		key, value := it.Element()
		i := b.dynamicIteration.MakeChild(spec.iteratorName, key, value)

		block, blockDiags := spec.newBlock(i, b.ctx)
		diags = append(diags, blockDiags...)
		if block != nil {
			// Attach our new iteration context so that attributes
			// and other nested blocks can refer to our iterator.
			block.Body = b.expandChild(block.Body, i, b.metaArgIteration)
			blocks = append(blocks, block)
		}
	}
	return blocks, diags
}

func (b *expandBody) expandMetaArgBlock(schema *hcl.BodySchema, rawBlock *hcl.Block) (hcl.Blocks, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if _, hidden := b.hiddenBlocks[rawBlock.Type]; hidden {
		return hcl.Blocks{}, diags
	}

	spec, specDiags := b.decodeMetaArgSpec(rawBlock)
	diags = append(diags, specDiags...)
	if specDiags.HasErrors() {
		return hcl.Blocks{}, diags
	}

	//// count attribute

	if spec.countSet {
		if !spec.countVal.IsKnown() {
			// If count is unknown, no blocks are returned
			return hcl.Blocks{}, diags
		}

		var blocks hcl.Blocks

		for idx := 0; idx < spec.countNum; idx++ {
			i := MakeCountIteration(cty.NumberIntVal(int64(idx)))

			expandedBlock := *rawBlock // shallow copy
			expandedBlock.Body = b.expandChild(rawBlock.Body, b.dynamicIteration, i)
			blocks = append(blocks, &expandedBlock)
		}

		return blocks, diags
	}

	//// for_each attribute

	if spec.forEachSet {
		if !spec.forEachVal.IsKnown() {
			// If for_each is unknown, no blocks are returned
			return hcl.Blocks{}, diags
		}

		var blocks hcl.Blocks

		for it := spec.forEachVal.ElementIterator(); it.Next(); {
			i := MakeForEachIteration(it.Element())

			expandedBlock := *rawBlock // shallow copy
			expandedBlock.Body = b.expandChild(rawBlock.Body, b.dynamicIteration, i)
			blocks = append(blocks, &expandedBlock)
		}

		return blocks, diags
	}

	//// Neither count/for_each

	return hcl.Blocks{b.expandStaticBlock(rawBlock)}, diags
}

func (b *expandBody) expandStaticBlock(rawBlock *hcl.Block) *hcl.Block {
	// A static block doesn't create a new iteration context, but
	// it does need to inherit _our own_ iteration context in
	// case it contains expressions that refer to our inherited
	// iterators, or nested "dynamic" blocks.
	expandedBlock := *rawBlock
	expandedBlock.Body = b.expandChild(rawBlock.Body, b.dynamicIteration, b.metaArgIteration)
	return &expandedBlock
}

func (b *expandBody) expandChild(child hcl.Body, i *dynamicIteration, mi *metaArgIteration) hcl.Body {
	chiCtx := i.EvalContext(mi.EvalContext(b.ctx))
	ret := Expand(child, chiCtx)
	ret.(*expandBody).dynamicIteration = i
	ret.(*expandBody).metaArgIteration = mi
	return ret
}

func (b *expandBody) JustAttributes() (hcl.Attributes, hcl.Diagnostics) {
	// blocks aren't allowed in JustAttributes mode and this body can
	// only produce blocks, so we'll just pass straight through to our
	// underlying body here.
	return b.original.JustAttributes()
}

func (b *expandBody) MissingItemRange() hcl.Range {
	return b.original.MissingItemRange()
}
