package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type Module struct {
	Resources   hclext.Blocks
	Variables   map[string]*Variable
	ModuleCalls map[string]*ModuleCall

	SourceDir string

	Sources map[string][]byte
	Files   map[string]*hcl.File

	primaries []*hcl.File
	overrides []*hcl.File
}

func NewEmptyModule() *Module {
	return &Module{
		Resources:   hclext.Blocks{},
		Variables:   map[string]*Variable{},
		ModuleCalls: map[string]*ModuleCall{},

		SourceDir: "",

		Sources: map[string][]byte{},
		Files:   map[string]*hcl.File{},

		primaries: []*hcl.File{},
		overrides: []*hcl.File{},
	}
}

func (m *Module) build() hcl.Diagnostics {
	body, diags := m.PartialContent(moduleSchema)
	if diags.HasErrors() {
		return diags
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			m.Resources = append(m.Resources, block)
		case "variable":
			v, valDiags := decodeVairableBlock(block)
			diags = diags.Extend(valDiags)
			m.Variables[v.Name] = v
		case "module":
			call, moduleDiags := decodeModuleBlock(block)
			diags = diags.Extend(moduleDiags)
			m.ModuleCalls[call.Name] = call
		}
	}

	return diags
}

func (m *Module) PartialContent(schema *hclext.BodySchema) (*hclext.BodyContent, hcl.Diagnostics) {
	content := &hclext.BodyContent{}
	diags := hcl.Diagnostics{}

	for _, f := range m.primaries {
		c, d := hclext.PartialContent(f.Body, schema)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = append(content.Blocks, c.Blocks...)
	}
	for _, f := range m.overrides {
		c, d := hclext.PartialContent(f.Body, schema)
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
			Body:       resourceBlockSchema,
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
	},
}
