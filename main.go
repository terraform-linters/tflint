package main

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/rules"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &terraform.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    "terraform",
				Version: "0.1.0",
			},
			PresetRules: map[string][]tflint.Rule{
				"all": {
					rules.NewTerraformCommentSyntaxRule(),
					rules.NewTerraformDeprecatedIndexRule(),
					rules.NewTerraformDeprecatedInterpolationRule(),
					rules.NewTerraformDocumentedOutputsRule(),
					rules.NewTerraformDocumentedVariablesRule(),
					rules.NewTerraformEmptyListEqualityRule(),
					rules.NewTerraformModulePinnedSourceRule(),
					rules.NewTerraformModuleVersionRule(),
					rules.NewTerraformNamingConventionRule(),
					rules.NewTerraformRequiredProvidersRule(),
					rules.NewTerraformRequiredVersionRule(),
					rules.NewTerraformStandardModuleStructureRule(),
					rules.NewTerraformTypedVariablesRule(),
					rules.NewTerraformUnusedDeclarationsRule(),
					rules.NewTerraformUnusedRequiredProvidersRule(),
					rules.NewTerraformWorkspaceRemoteRule(),
				},
				"recommended": {
					rules.NewTerraformDeprecatedIndexRule(),
					rules.NewTerraformDeprecatedInterpolationRule(),
					rules.NewTerraformEmptyListEqualityRule(),
					rules.NewTerraformModulePinnedSourceRule(),
					rules.NewTerraformModuleVersionRule(),
					rules.NewTerraformRequiredProvidersRule(),
					rules.NewTerraformRequiredVersionRule(),
					rules.NewTerraformTypedVariablesRule(),
					rules.NewTerraformUnusedDeclarationsRule(),
					rules.NewTerraformWorkspaceRemoteRule(),
				},
			},
		},
	})
}
