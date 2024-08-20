package tflint

import "github.com/hashicorp/hcl/v2"

// TextNode represents a text with range in the source code.
type TextNode struct {
	Bytes []byte
	Range hcl.Range
}
