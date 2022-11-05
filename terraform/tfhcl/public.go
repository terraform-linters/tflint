package tfhcl

import "github.com/hashicorp/hcl/v2"

func Expand(body hcl.Body, ctx *hcl.EvalContext) hcl.Body {
	return &expandBody{
		original:   body,
		forEachCtx: ctx,
	}
}
