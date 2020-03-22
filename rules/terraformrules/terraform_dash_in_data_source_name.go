package terraformrules

import (
	"fmt"
	"log"
	"strings"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformDashInDataSourceNameRule checks whether resources have any dashes in the name
type TerraformDashInDataSourceNameRule struct{}

// NewTerraformDashInDataSourceNameRule returns a new rule
func NewTerraformDashInDataSourceNameRule() *TerraformDashInDataSourceNameRule {
	return &TerraformDashInDataSourceNameRule{}
}

// Name returns the rule name
func (r *TerraformDashInDataSourceNameRule) Name() string {
	return "terraform_dash_in_data_source_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDashInDataSourceNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformDashInDataSourceNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDashInDataSourceNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether resources have any dashes in the name
func (r *TerraformDashInDataSourceNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, dataSource := range runner.TFConfig.Module.DataResources {
		if strings.Contains(dataSource.Name, "-") {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` data source name has a dash", dataSource.Name),
				dataSource.DeclRange,
			)
		}
	}

	return nil
}
