package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformAutofixRemoveLocal checks whether ...
type TerraformAutofixRemoveLocal struct {
	tflint.DefaultRule
}

// NewTerraformAutofixRemoveLocalRule returns a new rule
func NewTerraformAutofixRemoveLocalRule() *TerraformAutofixRemoveLocal {
	return &TerraformAutofixRemoveLocal{}
}

// Name returns the rule name
func (r *TerraformAutofixRemoveLocal) Name() string {
	return "terraform_autofix_remove_local"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformAutofixRemoveLocal) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformAutofixRemoveLocal) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TerraformAutofixRemoveLocal) Link() string {
	return ""
}

// Check checks whether ...
func (r *TerraformAutofixRemoveLocal) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	diags := hcl.Diagnostics{}
	for _, file := range files {
		content, _, schemaDiags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "locals"}},
		})
		diags = diags.Extend(schemaDiags)
		if schemaDiags.HasErrors() {
			continue
		}

		for _, block := range content.Blocks {
			attrs, localsDiags := block.Body.JustAttributes()
			diags = diags.Extend(localsDiags)
			if localsDiags.HasErrors() {
				continue
			}

			for name, attr := range attrs {
				if name == "autofix_removed" {
					if err := runner.EmitIssueWithFix(
						r,
						`Do not use "autofix_removed" local value`,
						attr.Range,
						func(f tflint.Fixer) error {
							return f.RemoveAttribute(attr)
						},
					); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
