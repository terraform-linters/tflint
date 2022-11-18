package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type Resource struct {
	Name string
	Type string

	DeclRange hcl.Range
	TypeRange hcl.Range
}

func decodeResourceBlock(block *hclext.Block) *Resource {
	return &Resource{
		Type:      block.Labels[0],
		Name:      block.Labels[1],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}
}
