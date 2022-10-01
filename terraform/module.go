package terraform

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

type Module struct {
	Resources   map[string]map[string]*Resource
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
		Resources:   map[string]map[string]*Resource{},
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
		}
	}

	return diags
}

// PartialContent extracts body content from Terraform configurations based on the passed schema.
// Basically, this function is a wrapper for hclext.PartialContent, but in some ways it reproduces
// Terraform language semantics.
//
//  1. The block schema implicitly adds dynamic blocks to the target
//     https://www.terraform.io/language/expressions/dynamic-blocks
//  2. Supports overriding files
//     https://www.terraform.io/language/files/override
//  3. Resources not created by count or for_each will be ignored
//     https://www.terraform.io/language/meta-arguments/count
//     https://www.terraform.io/language/meta-arguments/for_each
//
// But 3 won't run if you didn't pass the evaluation context.
func (m *Module) PartialContent(schema *hclext.BodySchema, ctx *Evaluator) (*hclext.BodyContent, hcl.Diagnostics) {
	content := &hclext.BodyContent{}
	diags := hcl.Diagnostics{}

	schema = schemaWithDynamic(schema)

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

	content = resolveDynamicBlocks(content)

	if ctx == nil {
		return content, diags
	}

	content, expandDiags := m.expandBlocks(content, ctx)
	diags = diags.Extend(expandDiags)

	return content, diags
}

// expandBlocks expands resource/module blocks depending on evaluation context.
// Currently, only decrementing block expansions, such as when count is 0 or for_each is empty,
// are supported, not incrementing expansions.
func (m *Module) expandBlocks(content *hclext.BodyContent, ctx *Evaluator) (*hclext.BodyContent, hcl.Diagnostics) {
	out := &hclext.BodyContent{Attributes: content.Attributes}
	diags := hcl.Diagnostics{}

	for _, block := range content.Blocks {
		switch block.Type {
		case "resource":
			resourceType := block.Labels[0]
			resourceName := block.Labels[1]

			resource := m.Resources[resourceType][resourceName]
			evaluable, evalDiags := ctx.ResourceIsEvaluable(resource)
			if evalDiags.HasErrors() {
				diags = diags.Extend(evalDiags)
				continue
			}

			if !evaluable {
				log.Printf("[WARN] Skip walking `%s` because it may not be created", resourceType+"."+resourceName)
				continue
			}
		case "module":
			name := block.Labels[0]

			module := m.ModuleCalls[name]
			evaluable, evalDiags := ctx.ModuleCallIsEvaluable(module)
			if evalDiags.HasErrors() {
				diags = diags.Extend(evalDiags)
				continue
			}

			if !evaluable {
				log.Printf("[WARN] Skip walking `module.%s` because it may not be created", name)
				continue
			}
		}

		out.Blocks = append(out.Blocks, block)
	}

	return out, diags
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

// schemaWithDynamic appends a dynamic block schema to block schemes recursively.
// The content retrieved by the added schema is formatted by resolveDynamicBlocks in the same way as regular blocks.
func schemaWithDynamic(schema *hclext.BodySchema) *hclext.BodySchema {
	out := &hclext.BodySchema{Attributes: schema.Attributes}

	for _, block := range schema.Blocks {
		block.Body = schemaWithDynamic(block.Body)

		out.Blocks = append(out.Blocks, block, hclext.BlockSchema{
			Type:       "dynamic",
			LabelNames: []string{"type"},
			Body: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "content",
						Body: block.Body,
					},
				},
			},
		})
	}

	return out
}

// resolveDynamicBlocks formats the passed content based on the block schema added by schemaWithDynamic.
// This allows you to get all named blocks without being aware of the difference in the structure of the dynamic block.
func resolveDynamicBlocks(content *hclext.BodyContent) *hclext.BodyContent {
	out := &hclext.BodyContent{Attributes: content.Attributes}

	for _, block := range content.Blocks {
		block.Body = resolveDynamicBlocks(block.Body)

		if block.Type != "dynamic" {
			out.Blocks = append(out.Blocks, block)
		} else {
			for _, dynamicContent := range block.Body.Blocks {
				dynamicContent.Type = block.Labels[0]
				out.Blocks = append(out.Blocks, dynamicContent)
			}
		}
	}

	return out
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
