package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformRequiredVersionRule checks whether a terraform version has required_version attribute
type TerraformRequiredVersionRule struct{}

// NewTerraformRequiredVersionRule returns new rule with default attributes
func NewTerraformRequiredVersionRule() *TerraformRequiredVersionRule {
	return &TerraformRequiredVersionRule{}
}

// Name returns the rule name
func (r *TerraformRequiredVersionRule) Name() string {
	return "terraform_required_version"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredVersionRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformRequiredVersionRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredVersionRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check Checks whether required_version is set
func (r *TerraformRequiredVersionRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "terraform",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "required_version"}},
				},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		return diags
	}

	var exists bool

	for _, block := range body.Blocks {
		_, ok := block.Body.Attributes["required_version"]
		exists = exists || ok
	}

	if !exists {
		runner.EmitIssue(
			r,
			`terraform "required_version" attribute is required`,
			hcl.Range{},
		)
	}

	return nil
}
