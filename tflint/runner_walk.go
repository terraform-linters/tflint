package tflint

import (
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// WalkResourceAttributes searches for resources and passes the appropriate attributes to the walker function
func (r *Runner) WalkResourceAttributes(resource, attributeName string, walker func(*hcl.Attribute) error) error {
	for _, resource := range r.LookupResourcesByType(resource) {
		ok, err := r.willEvaluateResource(resource)
		if err != nil {
			return err
		}
		if !ok {
			log.Printf("[WARN] Skip walking `%s` because it may not be created", resource.Type+"."+resource.Name)
			continue
		}

		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name: attributeName,
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		if attribute, ok := body.Attributes[attributeName]; ok {
			log.Printf("[DEBUG] Walk `%s` attribute", resource.Type+"."+resource.Name+"."+attributeName)
			err := r.WithExpressionContext(attribute.Expr, func() error {
				return walker(attribute)
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WalkResourceBlocks walks all blocks of the passed resource and invokes the passed function
func (r *Runner) WalkResourceBlocks(resource, blockType string, walker func(*hcl.Block) error) error {
	for _, resource := range r.LookupResourcesByType(resource) {
		ok, err := r.willEvaluateResource(resource)
		if err != nil {
			return err
		}
		if !ok {
			log.Printf("[WARN] Skip walking `%s` because it may not be created", resource.Type+"."+resource.Name)
			continue
		}

		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: blockType,
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, block := range body.Blocks {
			log.Printf("[DEBUG] Walk `%s` block", resource.Type+"."+resource.Name+"."+blockType)
			err := walker(block)
			if err != nil {
				return err
			}
		}

		// Walk in the same way for dynamic blocks. Note that we are not expanding blocks.
		// Therefore, expressions that use iterator are unevaluable.
		dynBody, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type:       "dynamic",
					LabelNames: []string{"name"},
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		for _, block := range dynBody.Blocks {
			if len(block.Labels) == 1 && block.Labels[0] == blockType {
				body, _, diags = block.Body.PartialContent(&hcl.BodySchema{
					Blocks: []hcl.BlockHeaderSchema{
						{
							Type: "content",
						},
					},
				})
				if diags.HasErrors() {
					return diags
				}

				for _, block := range body.Blocks {
					log.Printf("[DEBUG] Walk dynamic `%s` block", resource.Type+"."+resource.Name+"."+blockType)
					err := walker(block)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// WalkExpressions visits all blocks that can contain expressions:
// resource, data, module, provider, locals, and output. It calls the walker
// function with every expression it encounters and halts if the walker
// returns an error.
func (r *Runner) WalkExpressions(walker func(hcl.Expression) error) error {
	visit := func(expr hcl.Expression) error {
		return r.WithExpressionContext(expr, func() error {
			return walker(expr)
		})
	}

	for _, resource := range r.TFConfig.Module.ManagedResources {
		if err := r.walkBody(resource.Config, visit); err != nil {
			return err
		}
	}
	for _, resource := range r.TFConfig.Module.DataResources {
		if err := r.walkBody(resource.Config, visit); err != nil {
			return err
		}
	}
	for _, module := range r.TFConfig.Module.ModuleCalls {
		if err := r.walkBody(module.Config, visit); err != nil {
			return err
		}
	}
	for _, provider := range r.TFConfig.Module.ProviderConfigs {
		if err := r.walkBody(provider.Config, visit); err != nil {
			return err
		}
	}
	for _, local := range r.TFConfig.Module.Locals {
		if err := visit(local.Expr); err != nil {
			return err
		}
	}
	for _, output := range r.TFConfig.Module.Outputs {
		if err := visit(output.Expr); err != nil {
			return err
		}
	}

	return nil
}

// walkBody visits all attributes and passes their expressions to the walker function.
// It recurses on nested blocks.
func (r *Runner) walkBody(b hcl.Body, walker func(hcl.Expression) error) error {
	body, ok := b.(*hclsyntax.Body)
	if !ok {
		return r.walkAttributes(b, walker)
	}

	for _, attr := range body.Attributes {
		if err := walker(attr.Expr); err != nil {
			return err
		}
	}

	for _, block := range body.Blocks {
		if err := r.walkBody(block.Body, walker); err != nil {
			return err
		}
	}

	return nil
}

// walkAttributes visits all attributes and passes their expressions to the walker function.
// It should be used only for non-HCL bodies (JSON) when distinguishing a block from an attribute
// is not possible without a schema.
func (r *Runner) walkAttributes(b hcl.Body, walker func(hcl.Expression) error) error {
	attrs, diags := b.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	for _, attr := range attrs {
		if err := walker(attr.Expr); err != nil {
			return err
		}
	}

	return nil
}
