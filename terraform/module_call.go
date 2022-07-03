package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type ModuleCall struct {
	Name          string
	SourceAddrRaw string

	Count   hcl.Expression
	ForEach hcl.Expression

	DeclRange hcl.Range
}

func decodeModuleBlock(block *hclext.Block) (*ModuleCall, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	mc := &ModuleCall{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	if attr, exists := block.Body.Attributes["source"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &mc.SourceAddrRaw)
		diags = diags.Extend(valDiags)
	}

	if attr, exists := block.Body.Attributes["count"]; exists {
		mc.Count = attr.Expr
	}

	if attr, exists := block.Body.Attributes["for_each"]; exists {
		mc.ForEach = attr.Expr
	}

	return mc, diags
}

var moduleBlockSchema = &hclext.BodySchema{
	Attributes: []hclext.AttributeSchema{
		{
			Name: "source",
		},
		{
			Name: "count",
		},
		{
			Name: "for_each",
		},
	},
}
