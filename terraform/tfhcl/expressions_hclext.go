// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfhcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

// ExpandExpressionsHCLExt is ExpandVariablesHCLExt which returns
// []hcl.Expression instead of []hcl.Traversal.
func ExpandExpressionsHCLExt(body hcl.Body, schema *hclext.BodySchema) []hcl.Expression {
	rootNode := WalkExpandExpressions(body)
	return walkExpressionsWithHCLExt(rootNode, schema)
}

func walkExpressionsWithHCLExt(node WalkExpressionsNode, schema *hclext.BodySchema) []hcl.Expression {
	exprs, children := node.Visit(extendSchema(asHCLSchema(schema)))

	if len(children) > 0 {
		childSchemas := childBlockTypes(schema)
		for _, child := range children {
			if childSchema, exists := childSchemas[child.BlockTypeName]; exists {
				exprs = append(exprs, walkExpressionsWithHCLExt(child.Node, childSchema.Body)...)
			}
		}
	}

	return exprs
}

// WalkExpandExpressions is dynblock.WalkExpandVariables for expressions.
func WalkExpandExpressions(body hcl.Body) WalkExpressionsNode {
	return WalkExpressionsNode{body: body}
}

type WalkExpressionsNode struct {
	body          hcl.Body
	blockTypeName string
}

type WalkExpressionsChild struct {
	BlockTypeName string
	Node          WalkExpressionsNode
}

// Visit returns the expressions required for any "dynamic" blocks
// directly in the body associated with this node, and also returns any child
// nodes that must be visited in order to continue the walk.
//
// Each child node has its associated block type name given in its BlockTypeName
// field, which the calling application should use to determine the appropriate
// schema for the content of each child node and pass it to the child node's
// own Visit method to continue the walk recursively.
func (n WalkExpressionsNode) Visit(schema *hcl.BodySchema) (exprs []hcl.Expression, children []WalkExpressionsChild) {
	extSchema := n.extendSchema(schema)
	container, _, _ := n.body.PartialContent(extSchema)
	if container == nil {
		return exprs, children
	}

	children = make([]WalkExpressionsChild, 0, len(container.Blocks))

	for name, attr := range container.Attributes {
		// Special case: Terraform Core allows bare identifiers in
		// lifecycle.ignore_changes. These are attribute paths, not
		// variable references. To avoid treating them as variables or
		// collecting function calls from them, skip collecting the
		// expression here. This keeps behavior aligned with Core.
		if n.blockTypeName == "lifecycle" && name == "ignore_changes" {
			continue
		}
		exprs = append(exprs, attr.Expr)
	}

	for _, block := range container.Blocks {
		switch block.Type {

		case "dynamic":
			blockTypeName := block.Labels[0]
			inner, _, _ := block.Body.PartialContent(variableDetectionInnerSchema)
			if inner == nil {
				continue
			}

			if attr, exists := inner.Attributes["for_each"]; exists {
				exprs = append(exprs, attr.Expr)
			}
			if attr, exists := inner.Attributes["labels"]; exists {
				exprs = append(exprs, attr.Expr)
			}

			for _, contentBlock := range inner.Blocks {
				// We only request "content" blocks in our schema, so we know
				// any blocks we find here will be content blocks. We require
				// exactly one content block for actual expansion, but we'll
				// be more liberal here so that callers can still collect
				// expressions from erroneous "dynamic" blocks.
				children = append(children, WalkExpressionsChild{
					BlockTypeName: blockTypeName,
					Node: WalkExpressionsNode{
						body:          contentBlock.Body,
						blockTypeName: blockTypeName,
					},
				})
			}

		default:
			children = append(children, WalkExpressionsChild{
				BlockTypeName: block.Type,
				Node: WalkExpressionsNode{
					body:          block.Body,
					blockTypeName: block.Type,
				},
			})

		}
	}

	return exprs, children
}

func (c WalkExpressionsNode) extendSchema(schema *hcl.BodySchema) *hcl.BodySchema {
	// We augment the requested schema to also include our special "dynamic"
	// block type, since then we'll get instances of it interleaved with
	// all of the literal child blocks we must also include.
	extSchema := &hcl.BodySchema{
		Attributes: schema.Attributes,
		Blocks:     make([]hcl.BlockHeaderSchema, len(schema.Blocks), len(schema.Blocks)+1),
	}
	copy(extSchema.Blocks, schema.Blocks)
	extSchema.Blocks = append(extSchema.Blocks, dynamicBlockHeaderSchema)

	return extSchema
}

// This is a more relaxed schema than what's in schema.go, since we
// want to maximize the amount of variables we can find even if there
// are erroneous blocks.
var variableDetectionInnerSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "for_each",
			Required: false,
		},
		{
			Name:     "labels",
			Required: false,
		},
		{
			Name:     "iterator",
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "content",
		},
	},
}
