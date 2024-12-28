package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
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

	primaries         map[string]*hcl.File
	overrides         map[string]*hcl.File
	overrideFilenames []string
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

		primaries:         map[string]*hcl.File{},
		overrides:         map[string]*hcl.File{},
		overrideFilenames: []string{},
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
			v, valDiags := decodeVariableBlock(block)
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

// Rebuild rebuilds the module from the passed sources.
// The main purpose of this is to apply autofixes in the module.
func (m *Module) Rebuild(sources map[string][]byte) hcl.Diagnostics {
	if len(sources) == 0 {
		return nil
	}
	var diags hcl.Diagnostics

	for path, source := range sources {
		var file *hcl.File
		var d hcl.Diagnostics
		if strings.HasSuffix(path, ".json") {
			file, d = hcljson.Parse(source, path)
		} else {
			file, d = hclsyntax.ParseConfig(source, path, hcl.InitialPos)
		}
		if d.HasErrors() {
			diags = diags.Extend(d)
			continue
		}

		m.Sources[path] = source
		m.Files[path] = file
		if _, exists := m.primaries[path]; exists {
			m.primaries[path] = file
		}
		if _, exists := m.overrides[path]; exists {
			m.overrides[path] = file
		}
	}

	d := m.build()
	diags = diags.Extend(d)
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

	// Overrides are processed in order first by filename (in lexicographical order)
	for _, filename := range m.overrideFilenames {
		expanded, d := ctx.ExpandBlock(m.overrides[filename].Body, schema)
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

// blockAddr returns an identifier of the given block (e.g. "resource.aws_instance.main").
// This is used for overrides, which use the block type and labels as identifiers.
func blockAddr(b *hclext.Block) string {
	if len(b.Labels) > 0 {
		return fmt.Sprintf("%s.%s", b.Type, strings.Join(b.Labels, "."))
	}
	return b.Type
}

// overrideBlocks overrides the primary blocks passed with override blocks,
// following Terraform's merge behavior.
// https://developer.hashicorp.com/terraform/language/files/override#merging-behavior
//
// Note that this function returns the overwritten primary blocks
// but has side effects on the primary blocks and the overrides blocks.
func overrideBlocks(primaries, overrides hclext.Blocks) hclext.Blocks {
	overridesByAddr := map[string]hclext.Blocks{}
	for _, primary := range primaries {
		addr := blockAddr(primary)
		overridesByAddr[addr] = append(overridesByAddr[addr], primary)
	}

	// The block containing elements that cannot be overridden will be added as new primaries.
	// e.g. Local values ​not present in the primaries.
	//
	// Intuitively, if there is always only one corresponding block,
	// it is hard to imagine a case where it cannot be overwritten,
	// but since "locals" and "terraform" blocks can be declared multiple times,
	// please note that the block to be overwritten cannot be uniquely determined.
	newPrimaries := hclext.Blocks{}

	for _, override := range overrides {
		addr := blockAddr(override)

		switch override.Type {
		case "resource":
			if primaries, exists := overridesByAddr[addr]; exists {
				// Duplicate resource blocks are not allowed.
				overrideResourceBlock(primaries[0], override)
			}

		// The "data" block is the same as generic block except for "depends_on".
		// The "depends_on" arguments should not be merged, but Terraform will throw an error about it,
		// so we won't take that into consideration here.

		// The "variable" block is the same as generic block except for "type" and "default".
		// Conversion of default values ​​is done during evaluation and is not considered here.

		// The "output" block is the same as generic block except for "depends_on".

		case "locals":
			remain := overrideLocalBlocks(overridesByAddr[addr], override)
			if remain != nil {
				newPrimaries = append(newPrimaries, remain)
			}

		case "terraform":
			remain := overrideTerraformBlocks(overridesByAddr[addr], override)
			if remain != nil {
				newPrimaries = append(newPrimaries, remain)
			}

		default:
			if primaries, exists := overridesByAddr[addr]; exists {
				// The general rule, duplicated blocks are not allowed.
				overrideGenericBlock(primaries[0], override)
			}
		}
	}

	return append(primaries, newPrimaries...)
}

// overrideResourceBlock overrides "resource" block
// https://developer.hashicorp.com/terraform/language/files/override#merging-resource-and-data-blocks
//
// The "depends_on" arguments should not be merged, but Terraform will throw an error about it,
// so we won't take that into consideration here.
//
// This function modifies the given primary directly.
func overrideResourceBlock(primary, override *hclext.Block) {
	// An attribute argument within an override block
	// replaces any argument of the same name in the original block.
	for name, attr := range override.Body.Attributes {
		primary.Body.Attributes[name] = attr
	}

	// Exit early if blocks are empty.
	if len(primary.Body.Blocks) == 0 && len(override.Body.Blocks) == 0 {
		return
	}
	overridesByType := override.Body.Blocks.ByType()

	// Any nested blocks within an override block replace all blocks of the same type in the original block.
	// Any block types that do not appear in the override block remain from the original block.
	primary.Body.Blocks = filterBlocks(primary.Body.Blocks, func(p *hclext.Block) bool {
		overrides, exists := overridesByType[p.Type]
		if !exists {
			return true
		}

		if p.Type == "lifecycle" {
			// Contents of any lifecycle nested block are merged on an argument-by-argument basis.
			// Can't override nested blocks like precondition/postcondition.
			for _, override := range overrides {
				for name, attr := range override.Body.Attributes {
					p.Body.Attributes[name] = attr
				}
			}
			return true
		}

		return false
	})
	primary.Body.Blocks = append(
		primary.Body.Blocks,
		filterBlocks(override.Body.Blocks, func(b *hclext.Block) bool { return b.Type != "lifecycle" })...,
	)
}

// overrideLocalBlocks overrides "local" blocks
// https://developer.hashicorp.com/terraform/language/files/override#merging-locals-blocks
//
// This function modifies the given primaries directly.
// If the given override contains elements that cannot be overridden, (e.g. new local values)
// it is returned to the caller with only those elements remaining.
// This operation modifies the given override directly.
func overrideLocalBlocks(primaries hclext.Blocks, override *hclext.Block) *hclext.Block {
	// When there are multiple locals blocks,
	// it is not obvious into which one the remaining local values ​​should be merged.
	appendRemains := len(primaries) > 1

	// Tracks locals ​​that were not used to override.
	remains := hclext.Attributes{}
	for name, attr := range override.Body.Attributes {
		remains[name] = attr
	}

	// Overrides are applied on a value-by-value basis, ignoring which locals block they are defined in.
	for _, primary := range primaries {
		for name, attr := range override.Body.Attributes {
			// Track the remaining local values ​​only if you need to append to them,
			// otherwise simply merge them.
			if appendRemains {
				if _, exists := primary.Body.Attributes[name]; exists {
					primary.Body.Attributes[name] = attr
					delete(remains, name)
				}
			} else {
				primary.Body.Attributes[name] = attr
			}
		}
	}

	// Any remaining locals that aren't overridden will be added as a new block.
	if appendRemains && len(remains) > 0 {
		override.Body.Attributes = remains
		return override
	}
	return nil
}

// overrideTerraformBlocks overrides "terraform" blocks
// https://developer.hashicorp.com/terraform/language/files/override#merging-terraform-blocks
//
// This function modifies the given primaries directly.
// If the given override contains elements that cannot be overridden, (e.g. new required providers)
// it is returned to the caller with only those elements remaining.
// This operation modifies the given override directly.
func overrideTerraformBlocks(primaries hclext.Blocks, override *hclext.Block) *hclext.Block {
	// When there are multiple required_providers blocks,
	// it is not obvious into which one the remaining require providers ​​should be merged.
	appendRemains := false
	requiredProviderSeen := false
	for _, primary := range primaries {
		switch len(primary.Body.Blocks.ByType()["required_providers"]) {
		case 0:
			continue

		case 1:
			// Found multiple terraform blocks with required_providers
			if requiredProviderSeen {
				appendRemains = true
				break
			}
			requiredProviderSeen = true

		default:
			// Found terraform block with multiple required_providers
			appendRemains = true
			break
		}
	}

	// Tracks required providers ​​that were not used to override.
	remainRequiredProviders := override.Body.Blocks.ByType()["required_providers"]

	for _, primary := range primaries {
		// An attribute argument within an override block
		// replaces any argument of the same name in the original block.
		for name, attr := range override.Body.Attributes {
			primary.Body.Attributes[name] = attr
		}

		// Exit early if blocks are empty.
		if len(primary.Body.Blocks) == 0 && len(override.Body.Blocks) == 0 {
			continue
		}
		overridesByType := override.Body.Blocks.ByType()

		// Any nested blocks within an override block replace all blocks of the same type in the original block.
		// Any block types that do not appear in the override block remain from the original block.
		primary.Body.Blocks = filterBlocks(primary.Body.Blocks, func(p *hclext.Block) bool {
			switch p.Type {
			case "required_providers":
				// If the required_providers argument is set, its value is merged on an element-by-element basis
				for _, override := range overridesByType[p.Type] {
					for name, attr := range override.Body.Attributes {
						// Track the remaining required providers ​​only if you need to append to them,
						// otherwise simply merge them.
						if appendRemains {
							if _, exists := p.Body.Attributes[name]; exists {
								p.Body.Attributes[name] = attr
								for _, remain := range remainRequiredProviders {
									delete(remain.Body.Attributes, name)
								}
							}
						} else {
							p.Body.Attributes[name] = attr
						}
					}
				}
				return true

			case "cloud", "backend":
				// The presence of a block defining a backend (either cloud or backend) in an override file
				// always takes precedence over a block defining a backend in the original configuration.
				if _, exists := overridesByType["cloud"]; exists {
					return false
				}
				if _, exists := overridesByType["backend"]; exists {
					return false
				}
				return true

			default:
				_, exists := overridesByType[p.Type]
				return !exists
			}
		})
		primary.Body.Blocks = append(
			primary.Body.Blocks,
			filterBlocks(override.Body.Blocks, func(b *hclext.Block) bool { return b.Type != "required_providers" })...,
		)
	}

	// Any remaining required providers that aren't overridden will be added as a new block.
	if appendRemains {
		remainRequiredProviders = filterBlocks(remainRequiredProviders, func(b *hclext.Block) bool {
			return len(b.Body.Attributes) > 0
		})
		if len(remainRequiredProviders) > 0 {
			override.Body.Blocks = remainRequiredProviders
			return override
		}
	}
	return nil
}

// overrideGenericBlock overrides generic blocks.
// https://developer.hashicorp.com/terraform/language/files/override#merging-behavior
//
// Except for a few special blocks, most blocks are overridden by this rule.
// This function modifies the given primary directly.
func overrideGenericBlock(primary, override *hclext.Block) {
	// An attribute argument within an override block
	// replaces any argument of the same name in the original block.
	for name, attr := range override.Body.Attributes {
		primary.Body.Attributes[name] = attr
	}

	// Exit early if blocks are empty.
	if len(primary.Body.Blocks) == 0 && len(override.Body.Blocks) == 0 {
		return
	}
	overridesByType := override.Body.Blocks.ByType()

	// Any nested blocks within an override block replace all blocks of the same type in the original block.
	// Any block types that do not appear in the override block remain from the original block.
	primary.Body.Blocks = filterBlocks(primary.Body.Blocks, func(p *hclext.Block) bool {
		_, exists := overridesByType[p.Type]
		return !exists
	})
	primary.Body.Blocks = append(primary.Body.Blocks, override.Body.Blocks...)
}

func filterBlocks(in hclext.Blocks, fn func(*hclext.Block) bool) hclext.Blocks {
	out := hclext.Blocks{}
	for _, block := range in {
		if fn(block) {
			out = append(out, block)
		}
	}
	return out
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
