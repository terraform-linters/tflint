package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// LocalsJustAttributesExampleRule checks whether ...
type LocalsJustAttributesExampleRule struct {
	tflint.DefaultRule
}

// LocalsJustAttributesExampleRule returns a new rule
func NewLocalsJustAttributesExampleRule() *LocalsJustAttributesExampleRule {
	return &LocalsJustAttributesExampleRule{}
}

// Name returns the rule name
func (r *LocalsJustAttributesExampleRule) Name() string {
	return "locals_just_attributes_example"
}

// Enabled returns whether the rule is enabled by default
func (r *LocalsJustAttributesExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *LocalsJustAttributesExampleRule) Severity() tflint.Severity {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *LocalsJustAttributesExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *LocalsJustAttributesExampleRule) Check(runner tflint.Runner) error {
	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "locals",
				Body: &hclext.BodySchema{Mode: hclext.SchemaJustAttributesMode},
			},
		},
	}, nil)
	if err != nil || len(body.Blocks) == 0 {
		return err
	}

	locals := body.Blocks[0]

	if _, exists := locals.Body.Attributes["just_attributes"]; !exists {
		return nil
	}

	return runner.EmitIssue(r, fmt.Sprintf("found %d local values", len(locals.Body.Attributes)), locals.DefRange)
}
