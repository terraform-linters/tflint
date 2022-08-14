package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformDocumentedOutputsRule checks whether outputs have descriptions
type TerraformDocumentedOutputsRule struct {
	tflint.DefaultRule
}

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
	return project.ReferenceLink(r.Name())
}

// Check checks whether outputs have descriptions
func (r *TerraformDocumentedOutputsRule) Check(runner tflint.Runner) error {
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
				Type:       "output",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "description"}},
				},
			},
		},
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return err
	}

	for _, output := range body.Blocks {
		attr, exists := output.Body.Attributes["description"]
		if !exists {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` output has no description", output.Labels[0]),
				output.DefRange,
			); err != nil {
				return err
			}
			continue
		}

		var description string
		diags := gohcl.DecodeExpression(attr.Expr, nil, &description)
		if diags.HasErrors() {
			return diags
		}

		if description == "" {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` output has no description", output.Labels[0]),
				output.DefRange,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
