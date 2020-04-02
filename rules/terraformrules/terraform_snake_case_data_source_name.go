package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformSnakeCaseDataSourceNameRule checks whether data source names are snake case
type TerraformSnakeCaseDataSourceNameRule struct{}

// NewTerraformSnakeCaseDataSourceNameRule returns a new rule
func NewTerraformSnakeCaseDataSourceNameRule() *TerraformSnakeCaseDataSourceNameRule {
	return &TerraformSnakeCaseDataSourceNameRule{}
}

// Name returns the rule name
func (r *TerraformSnakeCaseDataSourceNameRule) Name() string {
	return "terraform_snake_case_data_source_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformSnakeCaseDataSourceNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformSnakeCaseDataSourceNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformSnakeCaseDataSourceNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether data source names are snake case
func (r *TerraformSnakeCaseDataSourceNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, dataSource := range runner.TFConfig.Module.DataResources {
	  isSnakeCase, _ := regexp.MatchString("^[a-z_]+$", dataSource.Name)
		if !isSnakeCase {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` data source name is not snake_case", dataSource.Name),
				dataSource.DeclRange,
			)
		}
	}

	return nil
}
