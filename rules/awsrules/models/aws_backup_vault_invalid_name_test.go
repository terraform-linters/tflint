// This file generated by `tools/model-rule-gen/main.go`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_AwsBackupVaultInvalidNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_backup_vault" "foo" {
	name = "example_backup_vault_example_backup_vault_example_backup_vault"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsBackupVaultInvalidNameRule(),
					Message: `name does not match valid pattern ^[a-zA-Z0-9\-\_\.]{1,50}$`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_backup_vault" "foo" {
	name = "example_backup_vault"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsBackupVaultInvalidNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues())
	}
}
