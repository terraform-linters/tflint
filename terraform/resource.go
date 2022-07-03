package terraform

import "github.com/terraform-linters/tflint-plugin-sdk/hclext"

var resourceBlockSchema = &hclext.BodySchema{
	Attributes: []hclext.AttributeSchema{
		{Name: "count"},
		{Name: "for_each"},
	},
}
