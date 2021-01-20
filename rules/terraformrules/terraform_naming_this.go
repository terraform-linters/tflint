package terraformrules

import (
	"fmt"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformNamingThisRule checks whether blocks follow naming convention
type TerraformNamingThisRule struct{}

// NewTerraformNamingThisRule returns new rule with default attributes
func NewTerraformNamingThisRule() *TerraformNamingThisRule {
	return &TerraformNamingThisRule{}
}

// Name returns the rule name
func (r *TerraformNamingThisRule) Name() string {
	return "terraform_naming_this"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformNamingThisRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformNamingThisRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformNamingThisRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether blocks follow naming convention
func (r *TerraformNamingThisRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}
	managedResources := runner.TFConfig.Module.ManagedResources

	// Group resources
	countPerType := make(map[string]int)
	for _, resource := range managedResources {
		countPerType[resource.Type]++
	}

	// Find resources with wrong name
	for _, resource := range managedResources {
		amount, _ := countPerType[resource.Type]

		if amount == 1 && resource.Name != "this" {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Found only one resource of type `%s`, therefore the resource name should be `this` but was `%s`", resource.Type, resource.Name),
				resource.DeclRange,
			)
		}
	}

	return nil
}
