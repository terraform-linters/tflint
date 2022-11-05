package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type Module struct {
	Resources   map[string]map[string]*Resource
	Variables   map[string]*Variable
	Locals      map[string]*Local
	ModuleCalls map[string]*ModuleCall

	SourceDir string

	Sources map[string][]byte
	Files   map[string]*hcl.File

	primaries []*hcl.File
	overrides []*hcl.File
}

func NewEmptyModule() *Module {
	return &Module{
		Resources:   map[string]map[string]*Resource{},
		Variables:   map[string]*Variable{},
		Locals:      map[string]*Local{},
		ModuleCalls: map[string]*ModuleCall{},

		SourceDir: "",

		Sources: map[string][]byte{},
		Files:   map[string]*hcl.File{},

		primaries: []*hcl.File{},
		overrides: []*hcl.File{},
	}
}

func (m *Module) build() hcl.Diagnostics {
	body, diags := m.PartialContent(moduleSchema, nil)
	if diags.HasErrors() {
		return diags
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			r := decodeResourceBlock(block)
			if _, exists := m.Resources[r.Type]; !exists {
				m.Resources[r.Type] = map[string]*Resource{}
			}
			m.Resources[r.Type][r.Name] = r
		case "variable":
			v, valDiags := decodeVairableBlock(block)
			diags = diags.Extend(valDiags)
			m.Variables[v.Name] = v
		case "module":
			call, moduleDiags := decodeModuleBlock(block)
			diags = diags.Extend(moduleDiags)
			m.ModuleCalls[call.Name] = call
		case "locals":
			locals := decodeLocalsBlock(block)
			for _, local := range locals {
				m.Locals[local.Name] = local
			}
		}
	}

	return diags
}

// PartialContent extracts body content from Terraform configurations based on the passed schema.
// Basically, this function is a wrapper for hclext.PartialContent, but in some ways it reproduces
// Terraform language semantics.
//
//  1. Supports overriding files
//     https://developer.hashicorp.com/terraform/language/files/override
//  2. Expands "dynamic" blocks
//     https://developer.hashicorp.com/terraform/language/expressions/dynamic-blocks
//  3. Expands resource/module depends on the meta-arguments
//     https://developer.hashicorp.com/terraform/language/meta-arguments/count
//     https://developer.hashicorp.com/terraform/language/meta-arguments/for_each
//
// But 2 and 3 won't run if you didn't pass the evaluation context.
func (m *Module) PartialContent(schema *hclext.BodySchema, ctx *Evaluator) (*hclext.BodyContent, hcl.Diagnostics) {
	content := &hclext.BodyContent{}
	diags := hcl.Diagnostics{}

	for _, f := range m.primaries {
		expanded, d := ctx.ExpandBlock(f.Body, schema)
		diags = diags.Extend(d)
		c, d := hclext.PartialContent(expanded, schema)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = append(content.Blocks, c.Blocks...)
	}
	for _, f := range m.overrides {
		expanded, d := ctx.ExpandBlock(f.Body, schema)
		diags = diags.Extend(d)
		c, d := hclext.PartialContent(expanded, schema)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = overrideBlocks(content.Blocks, c.Blocks)
	}

	return content, diags
}

// overrideBlocks changes the attributes in the passed primary blocks by override blocks recursively.
func overrideBlocks(primaries, overrides hclext.Blocks) hclext.Blocks {
	dict := map[string]*hclext.Block{}
	for _, primary := range primaries {
		key := fmt.Sprintf("%s[%s]", primary.Type, strings.Join(primary.Labels, ","))
		dict[key] = primary
	}

	for _, override := range overrides {
		key := fmt.Sprintf("%s[%s]", override.Type, strings.Join(override.Labels, ","))
		if primary, exists := dict[key]; exists {
			for name, attr := range override.Body.Attributes {
				primary.Body.Attributes[name] = attr
			}
			primary.Body.Blocks = overrideBlocks(primary.Body.Blocks, override.Body.Blocks)
		}
	}

	return primaries
}

var moduleSchema = &hclext.BodySchema{
	Blocks: []hclext.BlockSchema{
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
			Body:       variableBlockSchema,
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
			Body:       moduleBlockSchema,
		},
		{
			Type: "locals",
			Body: localBlockSchema,
		},
	},
}
