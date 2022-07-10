package terraformrules

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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
func (r *TerraformTypedVariablesRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformTypedVariablesRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether variables have type
func (r *TerraformTypedVariablesRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "type"}},
				},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		return diags
	}

	for _, variable := range body.Blocks {
		if _, exists := variable.Body.Attributes["type"]; !exists {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%v` variable has no type", variable.Labels[0]),
				variable.DefRange,
			)
		}
	}

	return nil
}
