package tfhcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

// ExpandVariablesHCLExt collects traversals (variables) from expressions
// within the given body according to the provided schema. It mirrors
// ExpandExpressionsHCLExt but returns the discovered traversals. This allows
// us to apply special-casing consistent with expression collection, such as
// skipping lifecycle.ignore_changes, where bare identifiers are not true
// variable references.
func ExpandVariablesHCLExt(body hcl.Body, schema *hclext.BodySchema) []hcl.Traversal {
	exprs := ExpandExpressionsHCLExt(body, schema)
	var result []hcl.Traversal
	for _, expr := range exprs {
		result = append(result, expr.Variables()...)
	}
	return result
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
