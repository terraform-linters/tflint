package terraformrules

import (
	"fmt"
	"log"
	"strings"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformDashInOutputNameRule checks whether outputs have any dashes in the name
type TerraformDashInOutputNameRule struct{}

// NewTerraformDashInOutputNameRule returns a new rule
func NewTerraformDashInOutputNameRule() *TerraformDashInOutputNameRule {
	return &TerraformDashInOutputNameRule{}
}

// Name returns the rule name
func (r *TerraformDashInOutputNameRule) Name() string {
	return "terraform_dash_in_output_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDashInOutputNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformDashInOutputNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDashInOutputNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether outputs have any dashes in the name
func (r *TerraformDashInOutputNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, output := range runner.TFConfig.Module.Outputs {
		if strings.Contains(output.Name, "-") {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` output name has a dash", output.Name),
				output.DeclRange,
			)
		}
	}

	return nil
}
