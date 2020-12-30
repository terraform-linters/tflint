package cmd

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-aws/aws"
	"github.com/terraform-linters/tflint-ruleset-aws/project"
	"github.com/terraform-linters/tflint-ruleset-aws/rules"
	"github.com/terraform-linters/tflint-ruleset-aws/rules/api"
)

func (cli *CLI) actAsAwsPlugin() int {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &aws.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    "aws",
				Version: fmt.Sprintf("%s-bundled", project.Version),
				Rules:   rules.Rules,
			},
			APIRules: api.Rules,
		},
	})

	return ExitCodeOK
}
