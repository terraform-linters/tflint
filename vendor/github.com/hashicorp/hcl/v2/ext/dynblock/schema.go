// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynblock

import "github.com/hashicorp/hcl/v2"

var dynamicBlockHeaderSchema = hcl.BlockHeaderSchema{
	Type:       "dynamic",
	LabelNames: []string{"type"},
}

var dynamicBlockBodySchemaLabels = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "for_each",
			Required: true,
		},
		{
			Name:     "iterator",
			Required: false,
		},
		{
			Name:     "labels",
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "content",
			LabelNames: nil,
		},
	},
}

var dynamicBlockBodySchemaNoLabels = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "for_each",
			Required: true,
		},
		{
			Name:     "iterator",
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "content",
			LabelNames: nil,
		},
	},
}
