package terraformrules

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint/terraform/configs"
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

	for _, required := range runner.TFConfig.Module.ProviderRequirements.RequiredProviders {
		r.checkProvider(runner, required)
	}

	return nil
}

func (r *TerraformUnusedRequiredProvidersRule) checkProvider(runner *tflint.Runner, required *configs.RequiredProvider) {
	for _, resource := range runner.TFConfig.Module.ManagedResources {
		if r.usesProvider(resource, required) {
			return
		}
	}

	for _, resource := range runner.TFConfig.Module.DataResources {
		if r.usesProvider(resource, required) {
			return
		}
	}

	for _, provider := range runner.TFConfig.Module.ProviderConfigs {
		if required.Name == provider.Name {
			return
		}
	}

	for _, module := range runner.TFConfig.Module.ModuleCalls {
		for _, provider := range module.Providers {
			if provider.InParent.Name == required.Name {
				return
			}
		}
	}

	runner.EmitIssue(
		r,
		fmt.Sprintf("provider '%s' is declared in required_providers but not used by the module", required.Name),
		required.DeclRange,
	)
}

func (r *TerraformUnusedRequiredProvidersRule) usesProvider(resource *configs.Resource, required *configs.RequiredProvider) bool {
	if resource.ProviderConfigRef != nil {
		return resource.ProviderConfigRef.Name == required.Name
	}

	return resource.Provider.Type == required.Name
}
