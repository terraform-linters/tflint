package cmd

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/rules"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

func (cli *CLI) actAsBundledPlugin() int {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &terraform.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    "terraform",
				Version: fmt.Sprintf("%s-bundled", project.Version),
			},
			PresetRules: rules.PresetRules,
		},
	})
	return ExitCodeOK
}
