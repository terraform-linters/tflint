package terraformrules

import (
	"fmt"
	"log"

	"github.com/zclconf/go-cty/cty"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformTypedVariablesRule checks whether variables have a type declared
type TerraformTypedVariablesRule struct{}

// NewTerraformTypedVariablesRule returns a new rule
func NewTerraformTypedVariablesRule() *TerraformTypedVariablesRule {
	return &TerraformTypedVariablesRule{}
}

// Name returns the rule name
func (r *TerraformTypedVariablesRule) Name() string {
	return "terraform_typed_variables"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformTypedVariablesRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformTypedVariablesRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformTypedVariablesRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether variables have type
func (r *TerraformTypedVariablesRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, variable := range runner.TFConfig.Module.Variables {
		if variable.Type == cty.DynamicPseudoType {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%v` variable has no type", variable.Name),
				variable.DeclRange,
			)
		}
	}

	return nil
}
