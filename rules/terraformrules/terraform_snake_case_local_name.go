package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformSnakeCaseLocalNameRule checks whether local names are snake case
type TerraformSnakeCaseLocalNameRule struct{}

// NewTerraformSnakeCaseLocalNameRule returns a new rule
func NewTerraformSnakeCaseLocalNameRule() *TerraformSnakeCaseLocalNameRule {
	return &TerraformSnakeCaseLocalNameRule{}
}

// Name returns the rule name
func (r *TerraformSnakeCaseLocalNameRule) Name() string {
	return "terraform_snake_case_local_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformSnakeCaseLocalNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformSnakeCaseLocalNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformSnakeCaseLocalNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether local names are snake case
func (r *TerraformSnakeCaseLocalNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, local := range runner.TFConfig.Module.Locals {
	  isSnakeCase, _ := regexp.MatchString("^[a-z_]+$", local.Name)
		if !isSnakeCase {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` local name is not snake_case", local.Name),
				local.DeclRange,
			)
		}
	}

	return nil
}
