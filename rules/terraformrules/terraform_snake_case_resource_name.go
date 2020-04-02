package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformSnakeCaseResourceNameRule checks whether resource names are snake case
type TerraformSnakeCaseResourceNameRule struct{}

// NewTerraformSnakeCaseResourceNameRule returns a new rule
func NewTerraformSnakeCaseResourceNameRule() *TerraformSnakeCaseResourceNameRule {
	return &TerraformSnakeCaseResourceNameRule{}
}

// Name returns the rule name
func (r *TerraformSnakeCaseResourceNameRule) Name() string {
	return "terraform_snake_case_resource_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformSnakeCaseResourceNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformSnakeCaseResourceNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformSnakeCaseResourceNameRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether resource names are snake case
func (r *TerraformSnakeCaseResourceNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, resource := range runner.TFConfig.Module.ManagedResources {
	  isSnakeCase, _ := regexp.MatchString("^[a-z_]+$", resource.Name)
		if !isSnakeCase {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` resource name is not snake_case", resource.Name),
				resource.DeclRange,
			)
		}
	}

	return nil
}
