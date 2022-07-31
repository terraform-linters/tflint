package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformBackendTypeRule checks whether ...
type TerraformBackendTypeRule struct {
	tflint.DefaultRule
}

// NewTerraformBackendTypeRule returns a new rule
func NewTerraformBackendTypeRule() *TerraformBackendTypeRule {
	return &TerraformBackendTypeRule{}
}

// Name returns the rule name
func (r *TerraformBackendTypeRule) Name() string {
	return "terraform_backend_type"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformBackendTypeRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformBackendTypeRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TerraformBackendTypeRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *TerraformBackendTypeRule) Check(runner tflint.Runner) error {
	// This rule is an example to get attributes of blocks other than resources.
	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "terraform",
				Body: &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "backend",
							LabelNames: []string{"type"},
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, terraform := range content.Blocks {
		for _, backend := range terraform.Body.Blocks {
			err := runner.EmitIssue(
				r,
				fmt.Sprintf("backend type is %s", backend.Labels[0]),
				backend.DefRange,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
