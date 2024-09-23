package terraform

import (
	"fmt"
	"maps"
	"slices"
	"sort"
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

	primaries map[string]*hcl.File
	overrides map[string]*hcl.File
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

		primaries: map[string]*hcl.File{},
		overrides: map[string]*hcl.File{},
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

	// If more than one override file defines the same top-level block, the overriding effect is compounded,
	// with later blocks taking precedence over earlier blocks.
	// Overrides are processed in order first by filename (in lexicographical order)
	// and then by position in each file.
	overrideFilenames := slices.Collect(maps.Keys(m.overrides))
	sort.Strings(overrideFilenames)
	for _, filename := range overrideFilenames {
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

// overrideBlocks overrides the primary blocks passed with override blocks,
// following Terraform's merge behavior.
// https://developer.hashicorp.com/terraform/language/files/override#merging-behavior
//
// Note that this function returns the overwritten primary blocks
// but has side effects on the primary blocks.
func overrideBlocks(primaries, overrides hclext.Blocks) hclext.Blocks {
	dict := map[string]hclext.Blocks{}
	for _, primary := range primaries {
		switch primary.Type {
		case "terraform":
			// The "terraform" blocks are allowed to be declared multiple times.
			dict[primary.Type] = append(dict[primary.Type], primary)

		default:
			// A top-level block in an override file merges with a block in a normal configuration file
			// that has the same block header.
			// The block header is the block type and any quoted labels that follow it.
			key := fmt.Sprintf("%s[%s]", primary.Type, strings.Join(primary.Labels, ","))
			dict[key] = hclext.Blocks{primary}
		}
	}

	newPrimaries := hclext.Blocks{}
	for _, override := range overrides {
		switch override.Type {
		case "terraform":
			// Any required_providers that were not used for overrides will be added,
			// so we will track whether they were used for overrides or not.
			overrideRequiredProviders := override.Body.Blocks.ByType()["required_providers"]

			for _, primary := range dict[override.Type] {
				// In both the required_version and required_providers settings,
				// each override constraint entirely replaces the constraints for
				// the same component in the original block.
				for name, attr := range override.Body.Attributes {
					primary.Body.Attributes[name] = attr
				}

				for _, overrideInnerBlock := range override.Body.Blocks {
					switch overrideInnerBlock.Type {
					case "required_providers":
						// If the required_providers argument is set, its value is merged on an element-by-element basis
						for _, primaryInnerBlock := range primary.Body.Blocks {
							if primaryInnerBlock.Type == "required_providers" {
								for name, attr := range overrideInnerBlock.Body.Attributes {
									if _, exists := primaryInnerBlock.Body.Attributes[name]; exists {
										primaryInnerBlock.Body.Attributes[name] = attr
										// Remove the required provider that was used to override.
										for _, requiredProvider := range overrideRequiredProviders {
											delete(requiredProvider.Body.Attributes, name)
										}
									}
								}
							}
						}

					case "cloud", "backend":
						// The presence of a block defining a backend (either cloud or backend) in an override file
						// always takes precedence over a block defining a backend in the original configuration.
						newInnerBlocks := hclext.Blocks{}
						for _, primaryInnerBlock := range primary.Body.Blocks {
							if primaryInnerBlock.Type != "cloud" && primaryInnerBlock.Type != "backend" {
								newInnerBlocks = append(newInnerBlocks, primaryInnerBlock)
							}
						}
						primary.Body.Blocks = append(newInnerBlocks, overrideInnerBlock)

					default:
						newInnerBlocks := hclext.Blocks{}
						for _, primaryInnerBlock := range primary.Body.Blocks {
							if primaryInnerBlock.Type != overrideInnerBlock.Type {
								newInnerBlocks = append(newInnerBlocks, primaryInnerBlock)
							}
						}
						primary.Body.Blocks = append(newInnerBlocks, overrideInnerBlock)
					}
				}
			}

			// Any remaining required providers that aren't overridden will be added as a new block.
			newRequiredProviders := hclext.Blocks{}
			for _, requiredProvider := range overrideRequiredProviders {
				if len(requiredProvider.Body.Attributes) > 0 {
					newRequiredProviders = append(newRequiredProviders, requiredProvider)
				}
			}
			if len(newRequiredProviders) > 0 {
				newPrimaries = append(newPrimaries, &hclext.Block{
					Type:   override.Type,
					Labels: override.Labels,
					Body: &hclext.BodyContent{
						Blocks: newRequiredProviders,
					},
					DefRange:    override.DefRange,
					TypeRange:   override.TypeRange,
					LabelRanges: override.LabelRanges,
				})
			}

		default:
			key := fmt.Sprintf("%s[%s]", override.Type, strings.Join(override.Labels, ","))
			if primaries, exists := dict[key]; exists {
				// The general rule, duplicated blocks are not allowed.
				primary := primaries[0]

				// Within a top-level block, an attribute argument within an override block
				// replaces any argument of the same name in the original block.
				for name, attr := range override.Body.Attributes {
					primary.Body.Attributes[name] = attr
				}

				// Within a top-level block, any nested blocks within an override block replace
				// all blocks of the same type in the original block.
				// Any block types that do not appear in the override block remain from the original block.
				for _, overrideInnerBlock := range override.Body.Blocks {
					newInnerBlocks := hclext.Blocks{}
					for _, primaryInnerBlock := range primary.Body.Blocks {
						if primaryInnerBlock.Type != overrideInnerBlock.Type {
							newInnerBlocks = append(newInnerBlocks, primaryInnerBlock)
						}
					}
					primary.Body.Blocks = append(newInnerBlocks, overrideInnerBlock)
				}
			}
		}
	}

	return append(primaries, newPrimaries...)
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
