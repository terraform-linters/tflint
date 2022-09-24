package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type Resource struct {
	Name    string
	Type    string
	Count   hcl.Expression
	ForEach hcl.Expression

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeResourceBlock(block *hclext.Block) *Resource {
	r := &Resource{
		Type:      block.Labels[0],
		Name:      block.Labels[1],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}

	if attr, exists := block.Body.Attributes["count"]; exists {
		r.Count = attr.Expr
	}
	if attr, exists := block.Body.Attributes["for_each"]; exists {
		r.ForEach = attr.Expr
	}
	return r
}

var resourceBlockSchema = &hclext.BodySchema{
	Attributes: []hclext.AttributeSchema{
		{Name: "count"},
		{Name: "for_each"},
	},
}
