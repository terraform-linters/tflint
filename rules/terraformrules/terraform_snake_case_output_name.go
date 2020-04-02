package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformSnakeCaseOutputNameRule checks whether output names are snake case
type TerraformSnakeCaseOutputNameRule struct{}

// NewTerraformSnakeCaseOutputNameRule returns a new rule
func NewTerraformSnakeCaseOutputNameRule() *TerraformSnakeCaseOutputNameRule {
	return &TerraformSnakeCaseOutputNameRule{}
}

// Name returns the rule name
func (r *TerraformSnakeCaseOutputNameRule) Name() string {
	return "terraform_snake_case_output_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformSnakeCaseOutputNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformSnakeCaseOutputNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformSnakeCaseOutputNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether output names are snake case
func (r *TerraformSnakeCaseOutputNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, output := range runner.TFConfig.Module.Outputs {
	  isSnakeCase, _ := regexp.MatchString("^[a-z_]+$", output.Name)
		if !isSnakeCase {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` output name is not snake_case", output.Name),
				output.DeclRange,
			)
		}
	}

	return nil
}
