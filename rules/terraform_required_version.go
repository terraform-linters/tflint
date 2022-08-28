package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformRequiredVersionRule checks whether a terraform version has required_version attribute
type TerraformRequiredVersionRule struct {
	tflint.DefaultRule
}

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
	return true
}

// Severity returns the rule severity
func (r *TerraformRequiredVersionRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredVersionRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check Checks whether required_version is set
func (r *TerraformRequiredVersionRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		// This rule does not run on non-Terraform directory.
		return nil
	}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "terraform",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "required_version"}},
				},
			},
		},
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return err
	}

	var exists bool

	for _, block := range body.Blocks {
		_, ok := block.Body.Attributes["required_version"]
		exists = exists || ok
	}

	if !exists {
		return runner.EmitIssue(
			r,
			`terraform "required_version" attribute is required`,
			hcl.Range{},
		)
	}

	return nil
}
