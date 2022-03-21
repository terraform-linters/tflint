package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformRequiredProvidersRule checks whether Terraform sets version constraints for all configured providers
type TerraformRequiredProvidersRule struct{}

// NewTerraformRequiredProvidersRule returns new rule with default attributes
func NewTerraformRequiredProvidersRule() *TerraformRequiredProvidersRule {
	return &TerraformRequiredProvidersRule{}
}

// Name returns the rule name
func (r *TerraformRequiredProvidersRule) Name() string {
	return "terraform_required_providers"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredProvidersRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformRequiredProvidersRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredProvidersRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check Checks whether provider required version is set
func (r *TerraformRequiredProvidersRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	providers := make(map[string]hcl.Range)
	module := runner.TFConfig.Module

	for _, provider := range module.ProviderConfigs {
		if _, ok := providers[provider.Name]; !ok {
			providers[provider.Name] = provider.DeclRange
		}

		if provider.Version.Required != nil {
			runner.EmitIssue(
				r,
				fmt.Sprintf(`%s: version constraint should be specified via "required_providers"`, provider.Addr().String()),
				provider.DeclRange,
			)
		}
	}

	providers = providerResourceRanges(providers, module.ManagedResources)
	providers = providerResourceRanges(providers, module.DataResources)

	for name, decl := range providers {
		if provider, ok := module.ProviderRequirements.RequiredProviders[name]; !ok || provider.Requirement.Required == nil {
			runner.EmitIssue(r, fmt.Sprintf(`Missing version constraint for provider "%s" in "required_providers"`, name), decl)
		}
	}

	return nil
}

func providerResourceRanges(providers map[string]hcl.Range, resources map[string]*configs.Resource) map[string]hcl.Range {
	for _, resource := range resources {
		provider := resource.Provider

		if provider.IsBuiltIn() {
			continue
		}

		if _, ok := providers[provider.Type]; ok {
			continue
		}

		providers[resource.Provider.Type] = resource.DeclRange
	}

	return providers
}
