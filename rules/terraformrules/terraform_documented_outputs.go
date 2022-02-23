package terraformrules

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformDocumentedOutputsRule checks whether outputs have descriptions
type TerraformDocumentedOutputsRule struct{}

// NewTerraformDocumentedOutputsRule returns a new rule
func NewTerraformDocumentedOutputsRule() *TerraformDocumentedOutputsRule {
	return &TerraformDocumentedOutputsRule{}
}

// Name returns the rule name
func (r *TerraformDocumentedOutputsRule) Name() string {
	return "terraform_documented_outputs"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDocumentedOutputsRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformDocumentedOutputsRule) Severity() tflint.Severity {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDocumentedOutputsRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether outputs have descriptions
func (r *TerraformDocumentedOutputsRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, output := range runner.TFConfig.Module.Outputs {
		if output.Description == "" {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` output has no description", output.Name),
				output.DeclRange,
			)
		}
	}

	return nil
}
