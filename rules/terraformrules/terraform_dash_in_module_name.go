package terraformrules

import (
	"fmt"
	"log"
	"strings"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformDashInModuleNameRule checks whether resources have any dashes in the name
type TerraformDashInModuleNameRule struct{}

// NewTerraformDashInModuleNameRule returns a new rule
func NewTerraformDashInModuleNameRule() *TerraformDashInModuleNameRule {
	return &TerraformDashInModuleNameRule{}
}

// Name returns the rule name
func (r *TerraformDashInModuleNameRule) Name() string {
	return "terraform_dash_in_module_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDashInModuleNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformDashInModuleNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDashInModuleNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether resources have any dashes in the name
func (r *TerraformDashInModuleNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, moduleCall := range runner.TFConfig.Module.ModuleCalls {
		if strings.Contains(moduleCall.Name, "-") {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` module name has a dash", moduleCall.Name),
				moduleCall.DeclRange,
			)
		}
	}

	return nil
}
