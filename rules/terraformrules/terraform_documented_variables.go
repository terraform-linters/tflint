package terraformrules

import (
	"fmt"
	"log"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/tflint"
)

// TerraformDocumentedVariablesRule checks whether variables have descriptions
type TerraformDocumentedVariablesRule struct{}

// NewTerraformDocumentedVariablesRule returns a new rule
func NewTerraformDocumentedVariablesRule() *TerraformDocumentedVariablesRule {
	return &TerraformDocumentedVariablesRule{}
}

// Name returns the rule name
func (r *TerraformDocumentedVariablesRule) Name() string {
	return "terraform_documented_variables"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDocumentedVariablesRule) Enabled() bool {
	return false
}

// Type returns the rule severity
func (r *TerraformDocumentedVariablesRule) Type() string {
	return issue.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDocumentedVariablesRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether variables have descriptions
func (r *TerraformDocumentedVariablesRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, variable := range runner.TFConfig.Module.Variables {
		if variable.Description == "" {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` variable has no description", variable.Name),
				variable.DeclRange,
			)
		}
	}

	return nil
}
