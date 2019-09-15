package terraformrules

import (
	"fmt"
	"log"
	"strings"

	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/tflint"
)

// TerraformDashInResourceNameRule checks whether resources have any dashes in the name
type TerraformDashInResourceNameRule struct{}

// NewTerraformDashInResourceNameRule returns a new rule
func NewTerraformDashInResourceNameRule() *TerraformDashInResourceNameRule {
	return &TerraformDashInResourceNameRule{}
}

// Name returns the rule name
func (r *TerraformDashInResourceNameRule) Name() string {
	return "terraform_dash_in_resource_name"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDashInResourceNameRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformDashInResourceNameRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformDashInResourceNameRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether resources have any dashes in the name
func (r *TerraformDashInResourceNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, resource := range runner.TFConfig.Module.ManagedResources {
		if strings.Contains(resource.Name, "-") {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%s` resource name has a dash", resource.Name),
				resource.DeclRange,
			)
		}
	}

	return nil
}
