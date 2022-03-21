package tflint

import (
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

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
		// HACK: Other than hclsyntax.Body, there are json.body and configs.mergeBody structs that satisfy hcl.Body,
		// but since both are private structs, there is no reliable way to determine them.
		// Here, it is judged by whether it can process `body.JustAttributes`. See also `walkAttributes`.
		if _, diags := b.JustAttributes(); diags.HasErrors() {
			log.Printf("[WARN] Ignore attributes of `%T` because we can only handle hclsyntax.Body or json.body", b)
			return nil
		}
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
