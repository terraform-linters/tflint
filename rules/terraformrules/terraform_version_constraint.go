package terraformrules

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
	"log"
)

// TerraformVersionConstraintRule checks whether a terraform version has
type TerraformVersionConstraintRule struct {
	attributeName string
}

type terraformVersionConstraintRuleConfig struct {
	Version string `hcl:"version,optional"`
}

// NewTerraformModulePinnedSourceRule returns new rule with default attributes
func NewTerraformVersionConstraintRule() *TerraformVersionConstraintRule {
	return &TerraformVersionConstraintRule{
		attributeName: "required_version",
	}
}

// Name returns the rule name
func (r *TerraformVersionConstraintRule) Name() string {
	return "terraform_version_constraint"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformVersionConstraintRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformVersionConstraintRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformVersionConstraintRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether variables have descriptions
func (r *TerraformVersionConstraintRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	module := runner.TFConfig.Module
	versionConstraints := module.CoreVersionConstraints
	if len(versionConstraints) == 0 {
		runner.EmitIssue(
			r,
			fmt.Sprintf("no terraform required_version attribute is declared"),
			hcl.Range{},
		)
		return nil
	}

	config := terraformVersionConstraintRuleConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	if config.Version == "" {
		return nil
	}

	for _, versionConstraint := range runner.TFConfig.Module.CoreVersionConstraints {
		if versionConstraint.Required.String() != config.Version {
			runner.EmitIssue(
				r,
				fmt.Sprintf("required_version does not match version \"%s\"", config.Version),
				versionConstraint.DeclRange,
			)
		}
	}

	return nil
}
