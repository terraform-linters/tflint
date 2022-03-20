package main

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/plugin/stub-generator/sources/testing/rules"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "testing",
			Version: "0.1.0",
			Rules: []tflint.Rule{
				rules.NewAwsAutoscalingGroupCtyEvalExampleRule(),
				rules.NewAwsInstanceExampleTypeRule(),
				rules.NewAwsS3BucketExampleLifecycleRuleRule(),
				rules.NewAwsInstanceMapEvalExampleRule(),
				rules.NewAwsS3BucketWithConfigExampleRule(),
			},
		},
	})
}
