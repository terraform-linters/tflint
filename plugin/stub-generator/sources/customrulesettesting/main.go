package main

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/plugin/stub-generator/sources/customrulesettesting/custom"
	"github.com/terraform-linters/tflint/plugin/stub-generator/sources/customrulesettesting/rules"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &custom.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    "customrulesettesting",
				Version: "0.1.0",
				Rules: []tflint.Rule{
					rules.NewAwsInstanceExampleTypeRule(),
				},
			},
		},
	})
}
