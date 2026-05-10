// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: BUSL-1.1

package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type ModuleCall struct {
	Name           string
	SourceExpr     hcl.Expression
	SourceResolved string

	DeclRange hcl.Range
}

func decodeModuleBlock(block *hclext.Block) (*ModuleCall, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	mc := &ModuleCall{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	if attr, exists := block.Body.Attributes["source"]; exists {
		mc.SourceExpr = attr.Expr
	}

	return mc, diags
}

var moduleBlockSchema = &hclext.BodySchema{
	Attributes: []hclext.AttributeSchema{
		{
			Name: "source",
		},
	},
}

// CallModuleType is a type of module to call.
// This is primarily used to control module walker behavior.
type CallModuleType int32

const (
	// CallAllModule calls all (local/remote) modules.
	CallAllModule CallModuleType = iota

	// CallLocalModule calls only local modules.
	CallLocalModule

	// CallNoModule does not call any modules.
	CallNoModule
)

func AsCallModuleType(s string) (CallModuleType, error) {
	switch s {
	case "all":
		return CallAllModule, nil
	case "local":
		return CallLocalModule, nil
	case "none":
		return CallNoModule, nil
	default:
		return CallAllModule, fmt.Errorf("%s is invalid call module type. Allowed values are: all, local, none", s)
	}
}

func (c CallModuleType) String() string {
	switch c {
	case CallAllModule:
		return "all"
	case CallLocalModule:
		return "local"
	case CallNoModule:
		return "none"
	default:
		panic("never happened")
	}
}
