package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformTypedVariablesRule checks whether variables have a type declared
type TerraformTypedVariablesRule struct {
	tflint.DefaultRule
}

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
	return true
}

// Severity returns the rule severity
func (r *TerraformTypedVariablesRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformTypedVariablesRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether variables have type
func (r *TerraformTypedVariablesRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "type"}},
				},
			},
		},
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return err
	}

	for _, variable := range body.Blocks {
		if _, exists := variable.Body.Attributes["type"]; !exists {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("`%v` variable has no type", variable.Labels[0]),
				variable.DefRange,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
