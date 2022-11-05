package tfhcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

// ExpandVariablesHCLExt is a wrapper around dynblock.WalkVariables that
// uses the given hclext.BodySchema to automatically drive the recursive
// walk through nested blocks in the given body.
//
// Note that it's a wrapper around ExpandVariables, not WalkExpandVariables.
// This package evaluates expressions immediately on expansion, so we always
// need all variables to expand. It also implicitly walks count/for_each to
// support expansion by meta-arguments.
func ExpandVariablesHCLExt(body hcl.Body, schema *hclext.BodySchema) []hcl.Traversal {
	rootNode := dynblock.WalkVariables(body)
	return walkVariablesWithHCLExt(rootNode, schema)
}

func walkVariablesWithHCLExt(node dynblock.WalkVariablesNode, schema *hclext.BodySchema) []hcl.Traversal {
	vars, children := node.Visit(extendSchema(asHCLSchema(schema)))

	if len(children) > 0 {
		childSchemas := childBlockTypes(schema)
		for _, child := range children {
			if childSchema, exists := childSchemas[child.BlockTypeName]; exists {
				vars = append(vars, walkVariablesWithHCLExt(child.Node, childSchema.Body)...)
			}
		}
	}

	return vars
}

func asHCLSchema(in *hclext.BodySchema) *hcl.BodySchema {
	out := &hcl.BodySchema{}
	if in == nil || in.Mode == hclext.SchemaJustAttributesMode {
		return out
	}

	out.Attributes = make([]hcl.AttributeSchema, len(in.Attributes))
	for idx, attr := range in.Attributes {
		out.Attributes[idx] = hcl.AttributeSchema{Name: attr.Name, Required: attr.Required}
	}
	out.Blocks = make([]hcl.BlockHeaderSchema, len(in.Blocks))
	for idx, block := range in.Blocks {
		out.Blocks[idx] = hcl.BlockHeaderSchema{Type: block.Type, LabelNames: block.LabelNames}
	}
	return out
}

func extendSchema(schema *hcl.BodySchema) *hcl.BodySchema {
	schema.Attributes = append(schema.Attributes, hcl.AttributeSchema{Name: "count"}, hcl.AttributeSchema{Name: "for_each"})
	return schema
}

func childBlockTypes(schema *hclext.BodySchema) map[string]hclext.BlockSchema {
	ret := make(map[string]hclext.BlockSchema)
	for _, block := range schema.Blocks {
		ret[block.Type] = block
	}
	return ret
}
