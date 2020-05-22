package terraformrules

import (
	"io/ioutil"
	"os"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformBlacklistedResourcesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: "empty configuration",
			Content: `
resource "random_id" "server" {
 byte_length = 8
}

resource "aws_iam_policy_attachment" "test-attach" {
 name       = "test-attachment"
 users      = [aws_iam_user.user.name]
 roles      = [aws_iam_role.role.name]
 groups     = [aws_iam_group.group.name]
 policy_arn = aws_iam_policy.policy.arn
}`,
			Config: `
rule "terraform_blacklisted_resources" {
 enabled = true
 types   = {
   google_organization_iam_binding = "This resource is banned from usage"
 }
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "with types",
			Content: `
resource "random_id" "server" {
  byte_length = 8
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "test-attachment"
  users      = [aws_iam_user.user.name]
  roles      = [aws_iam_role.role.name]
  groups     = [aws_iam_group.group.name]
  policy_arn = aws_iam_policy.policy.arn
}

resource "aws_iam_policy_attachment" "test-attach-2" {
  name       = "test-attachment-2"
  users      = [aws_iam_user.user.name]
  roles      = [aws_iam_role.role.name]
  groups     = [aws_iam_group.group.name]
  policy_arn = aws_iam_policy.policy.arn
}
`,
			Config: `
rule "terraform_blacklisted_resources" {
  enabled = true
  types   = {
    aws_iam_policy_attachment = "Consider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead."
  }
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformBlacklistedResourcesRule(),
					Message: "`aws_iam_policy_attachment` resource type is blacklisted\n\nConsider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead.",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 6, Column: 10},
						End:      hcl.Pos{Line: 6, Column: 37},
					},
				},
				{
					Rule:    NewTerraformBlacklistedResourcesRule(),
					Message: "`aws_iam_policy_attachment` resource type is blacklisted\n\nConsider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead.",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 14, Column: 10},
						End:      hcl.Pos{Line: 14, Column: 37},
					},
				},
			},
		},
		{
			Name: "with multiple types",
			Content: `
resource "random_id" "server" {
 byte_length = 8
}

resource "aws_iam_policy_attachment" "test-attach" {
 name       = "test-attachment"
 users      = [aws_iam_user.user.name]
 roles      = [aws_iam_role.role.name]
 groups     = [aws_iam_group.group.name]
 policy_arn = aws_iam_policy.policy.arn
}`,
			Config: `
rule "terraform_blacklisted_resources" {
 enabled = true
 types   = {
   aws_iam_policy_attachment = "Consider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead."
   random_id 				  = "Consider random_uuid as an alternative"
 }
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformBlacklistedResourcesRule(),
					Message: "`aws_iam_policy_attachment` resource type is blacklisted\n\nConsider aws_iam_role_policy_attachment, aws_iam_user_policy_attachment, or aws_iam_group_policy_attachment instead.",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 6, Column: 10},
						End:      hcl.Pos{Line: 6, Column: 37},
					},
				},
				{
					Rule:    NewTerraformBlacklistedResourcesRule(),
					Message: "`random_id` resource type is blacklisted\n\nConsider random_uuid as an alternative",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 2, Column: 10},
						End:      hcl.Pos{Line: 2, Column: 21},
					},
				},
			},
		},
	}

	rule := NewTerraformBlacklistedResourcesRule()

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"module.tf": tc.Content}, loadConfigFromBlacklistedResourceTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// TODO: Replace with TestRunner
func loadConfigFromBlacklistedResourceTempFile(t *testing.T, content string) *tflint.Config {
	if content == "" {
		return tflint.EmptyConfig()
	}

	tmpfile, err := ioutil.TempFile("", "terraform_blacklisted_resources")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	config, err := tflint.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return config
}
