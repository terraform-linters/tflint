package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformSnakeCaseVariableNameRule checks whether variable names are snake case
type TerraformSnakeCaseVariableNameRule struct{}

// NewTerraformSnakeCaseVariableNameRule returns a new rule
func NewTerraformSnakeCaseVariableNameRule() *TerraformSnakeCaseVariableNameRule {
	return &TerraformSnakeCaseVariableNameRule{}
}

// Name returns the rule name
func (r *TerraformSnakeCaseVariableNameRule) Name() string {
	return "terraform_snake_case_variable_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformSnakeCaseVariableNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformSnakeCaseVariableNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformSnakeCaseVariableNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether variable names are snake case
func (r *TerraformSnakeCaseVariableNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, variable := range runner.TFConfig.Module.Variables {
	  isSnakeCase, _ := regexp.MatchString("^[a-z_]+$", variable.Name)
		if !isSnakeCase {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` variable name is not snake_case", variable.Name),
				variable.DeclRange,
			)
		}
	}

	return nil
}
