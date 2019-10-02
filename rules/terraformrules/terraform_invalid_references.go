package terraformrules

import (
	"fmt"
	"log"

	"github.com/wata727/tflint/tflint"
)

// TerraformInvalidReferencesRule checks for variable references without braces
type TerraformInvalidReferencesRule struct{}

// NewTerraformInvalidReferencesRule returns a new rule
func NewTerraformInvalidReferencesRule() *TerraformInvalidReferencesRule {
	return &TerraformInvalidReferencesRule{}
}

// Name returns the rule name
func (r *TerraformInvalidReferencesRule) Name() string {
	return "terraform_invalid_references"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformInvalidReferencesRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformInvalidReferencesRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformInvalidReferencesRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check goes over variables and checks is they are referenced without brances
func (r *TerraformInvalidReferencesRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, variable := range runner.TFConfig.Module.Variables {
		// iterate over resources
		// find expandable strings
		// match variable against strings content

		// TODO: ManagedResources + DataResources
		for _, resource := range runner.TFConfig.Module.ManagedResources {
			fmt.Sprintf("`%s` resource", resource.Name)
		}
	}

	return nil
}
