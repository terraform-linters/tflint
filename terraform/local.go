package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type Local struct {
	Name string
	Expr hcl.Expression

	DeclRange hcl.Range
}

func decodeLocalsBlock(block *hclext.Block) []*Local {
	locals := make([]*Local, 0, len(block.Body.Attributes))
	for name, attr := range block.Body.Attributes {
		locals = append(locals, &Local{
			Name:      name,
			Expr:      attr.Expr,
			DeclRange: attr.Range,
		})
	}
	return locals
}

var localBlockSchema = &hclext.BodySchema{
	Mode: hclext.SchemaJustAttributesMode,
}
