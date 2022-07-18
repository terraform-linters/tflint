package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
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

// Severity returns the rule severity
func (r *TerraformDocumentedVariablesRule) Severity() tflint.Severity {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDocumentedVariablesRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether variables have descriptions
func (r *TerraformDocumentedVariablesRule) Check(runner *tflint.Runner) error {
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
					Attributes: []hclext.AttributeSchema{{Name: "description"}},
				},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		return diags
	}

	for _, variable := range body.Blocks {
		attr, exists := variable.Body.Attributes["description"]
		if !exists {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` variable has no description", variable.Labels[0]),
				variable.DefRange,
			)
			continue
		}

		var description string
		diags = gohcl.DecodeExpression(attr.Expr, nil, &description)
		if diags.HasErrors() {
			return diags
		}

		if description == "" {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` variable has no description", variable.Labels[0]),
				variable.DefRange,
			)
		}
	}

	return nil
}
