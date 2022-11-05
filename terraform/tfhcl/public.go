// Package tfhcl is a fork of hcl/ext/dynblock.
// Like dynblock, it supports dynamic block expansion, but also resource
// expansion via count/for_each meta-arguments.
// This package is defined separately from hclext because meta-arguments
// are a Terraform concern.
package tfhcl

import "github.com/hashicorp/hcl/v2"

// Expand "dynamic" blocks and count/for_for_each meta-arguments resources
// in the given body, returning a new body that has those blocks expanded.
//
// The given EvalContext is used when evaluating attributes within the given
// body. If the body has a dynamic block or an expandable resource, its
// contents are evaluated immediately.
//
// Expand returns no diagnostics because no blocks are actually expanded
// until a call to Content or PartialContent on the returned body, which
// will then expand only the blocks selected by the schema.
func Expand(body hcl.Body, ctx *hcl.EvalContext) hcl.Body {
	return &expandBody{
		original: body,
		ctx:      ctx,
	}
}
