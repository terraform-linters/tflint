// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynblock

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ExpandOption interface {
	applyExpandOption(*expandBody)
}

type optCheckForEach struct {
	check func(cty.Value, hcl.Expression, *hcl.EvalContext) hcl.Diagnostics
}

func OptCheckForEach(check func(cty.Value, hcl.Expression, *hcl.EvalContext) hcl.Diagnostics) ExpandOption {
	return optCheckForEach{check}
}

// applyExpandOption implements ExpandOption.
func (o optCheckForEach) applyExpandOption(body *expandBody) {
	body.checkForEach = append(body.checkForEach, o.check)
}
