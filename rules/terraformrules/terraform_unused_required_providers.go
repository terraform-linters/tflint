package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformUnusedRequiredProvidersRule checks whether required providers are used in the module
type TerraformUnusedRequiredProvidersRule struct{}

// NewTerraformUnusedRequiredProvidersRule returns new rule with default attributes
func NewTerraformUnusedRequiredProvidersRule() *TerraformUnusedRequiredProvidersRule {
	return &TerraformUnusedRequiredProvidersRule{}
}

// Name returns the rule name
func (r *TerraformUnusedRequiredProvidersRule) Name() string {
	return "terraform_unused_required_providers"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformUnusedRequiredProvidersRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformUnusedRequiredProvidersRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformUnusedRequiredProvidersRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether required providers are used
func (r *TerraformUnusedRequiredProvidersRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	providerRefs, diags := getProviderRefs(runner)
	if diags.HasErrors() {
		return diags
	}

	moduleRunners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return err
	}

	requiredProviders, diags := getRequiredProviders(runner)
	if diags.HasErrors() {
		return diags
	}

RequiredProvidersLoop:
	for key, required := range requiredProviders {
		if _, exists := providerRefs[required.Name]; !exists {
			for _, runner := range moduleRunners {
				moduleRequiredProviders, diags := getRequiredProviders(runner)
				if diags.HasErrors() {
					return diags
				}

				if _, exists := moduleRequiredProviders[key]; exists {
					continue RequiredProvidersLoop
				}
			}

			runner.EmitIssue(
				r,
				fmt.Sprintf("provider '%s' is declared in required_providers but not used by the module", required.Name),
				required.Range,
			)
		}
	}

	return nil
}

func getRequiredProviders(runner *tflint.Runner) (hcl.Attributes, hcl.Diagnostics) {
	requiredProviders := hcl.Attributes{}
	diags := hcl.Diagnostics{}

	for _, file := range runner.Files() {
		content, _, schemaDiags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "terraform"}},
		})
		diags = diags.Extend(schemaDiags)
		if diags.HasErrors() {
			continue
		}

		for _, block := range content.Blocks {
			content, _, schemaDiags = block.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{{Type: "required_providers"}},
			})
			diags = diags.Extend(schemaDiags)
			if diags.HasErrors() {
				continue
			}

			for _, block := range content.Blocks {
				var attrDiags hcl.Diagnostics
				requiredProviders, attrDiags = block.Body.JustAttributes()
				diags = diags.Extend(attrDiags)
				if diags.HasErrors() {
					continue
				}
			}
		}
	}
	if diags.HasErrors() {
		return nil, diags
	}

	return requiredProviders, nil
}
