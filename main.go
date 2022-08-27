package main

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/rules"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &terraform.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    "terraform",
				Version: project.Version,
			},
			PresetRules: rules.PresetRules,
		},
	})
}
