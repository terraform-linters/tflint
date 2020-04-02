package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformSnakeCaseModuleNameRule checks whether resource names are snake case
type TerraformSnakeCaseModuleNameRule struct{}

// NewTerraformSnakeCaseModuleNameRule returns a new rule
func NewTerraformSnakeCaseModuleNameRule() *TerraformSnakeCaseModuleNameRule {
	return &TerraformSnakeCaseModuleNameRule{}
}

// Name returns the rule name
func (r *TerraformSnakeCaseModuleNameRule) Name() string {
	return "terraform_snake_case_module_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformSnakeCaseModuleNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformSnakeCaseModuleNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformSnakeCaseModuleNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether resource names are snake case
func (r *TerraformSnakeCaseModuleNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, moduleCall := range runner.TFConfig.Module.ModuleCalls {
	  isSnakeCase, _ := regexp.MatchString("^[a-z_]+$", moduleCall.Name)
		if !isSnakeCase {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` module name is not snake_case", moduleCall.Name),
				moduleCall.DeclRange,
			)
		}
	}

	return nil
}
