package terraformrules

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformUnusedRequiredProvidersRule checks whether Terraform sets version constraints for all configured providers
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
func (r *TerraformUnusedRequiredProvidersRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformUnusedRequiredProvidersRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

//Check Checks whether provider required version is set
func (r *TerraformUnusedRequiredProvidersRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

providers:
	for _, required := range runner.TFConfig.Module.ProviderRequirements.RequiredProviders {
		for _, resource := range runner.TFConfig.Module.ManagedResources {
			if required.Name == resource.Provider.Type {
				continue providers
			}
		}

		for _, resource := range runner.TFConfig.Module.DataResources {
			if required.Name == resource.Provider.Type {
				continue providers
			}
		}

		for _, provider := range runner.TFConfig.Module.ProviderConfigs {
			if required.Name == provider.Name {
				continue providers
			}
		}

		runner.EmitIssue(
			r,
			fmt.Sprintf("provider '%s' is declared in required_providers but not used by the module", required.Name),
			required.DeclRange,
		)
	}

	return nil
}
