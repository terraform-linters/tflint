package main

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/rules"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "terraform",
			Version: "0.1.0",
			Rules: []tflint.Rule{
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
		},
	})
}
