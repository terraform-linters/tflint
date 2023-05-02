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
				rules.NewAwsIAMPolicyExampleRule(),
				rules.NewAwsInstanceExampleTypeRule(),
				rules.NewAwsS3BucketExampleLifecycleRuleRule(),
				rules.NewAwsInstanceMapEvalExampleRule(),
				rules.NewAwsS3BucketWithConfigExampleRule(),
				rules.NewAwsRoute53RecordEvalOnRootCtxExampleRule(),
				rules.NewAwsDBInstanceWithDefaultConfigExampleRule(),
				rules.NewAwsCloudFormationStackErrorRule(),
				rules.NewLocalsJustAttributesExampleRule(),
				rules.NewAwsIAMRoleExampleRule(),
				rules.NewTestingAssertionsExampleRule(),
				rules.NewTerraformAutofixRemoveLocalRule(), // should be former than terraform_autofix_comment because this rule changes the line number
				rules.NewTerraformAutofixCommentRule(),
				rules.NewAwsInstanceAutofixConflictRule(), // should be later than terraform_autofix_comment because this rule adds an issue for terraform_autofix_comment
			},
		},
	})
}
