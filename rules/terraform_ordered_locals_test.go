package rules

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformLocalsOrderRule(t *testing.T) {
	expectedIssue := &helper.Issue{
		Rule:    NewTerraformOrderedLocalsRule(),
		Message: "Local values must be in alphabetical order",
	}
	cases := []struct {
		Name     string
		JSON     bool
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "correct locals variable order",
			Content: `
locals {
  common_tags = {
    Service = local.service_name
    Owner   = local.owner
  }
  instance_ids = concat(aws_instance.blue.*.id, aws_instance.green.*.id)
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "empty locals block",
			Content: `
locals {}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "sorting in alphabetic order",
			Content: `
locals {
  instance_ids = concat(aws_instance.blue.*.id, aws_instance.green.*.id)
  common_tags = {
    Service = local.service_name
    Owner   = local.owner
  }
}`,
			Expected: helper.Issues{
				expectedIssue,
			},
		},
		{
			Name: "json",
			Content: `{
  "locals": [
    {
      "instance_ids": [
        "id1",
        "id2"
      ],
      "common_tags": [
        {
          "Owner": "Dev",
          "Service": "App"
        }
      ]
    }
  ]
}`,
			JSON: true,
			Expected: helper.Issues{
				expectedIssue,
			},
		},
		{
			Name: "multiple locals block in the same file",
			Content: `
locals {
  instance_ids = concat(aws_instance.blue.*.id, aws_instance.green.*.id)
  common_tags = {
    Service = local.service_name
    Owner   = local.owner
  }
}

locals {
  service_name = "forum"
  owner        = "Community Team"
}`,
			Expected: helper.Issues{
				expectedIssue,
				expectedIssue,
			},
		},
	}
	rule := NewTerraformOrderedLocalsRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			filename := "config.tf"
			if tc.JSON {
				filename = "config.tf.json"
			}
			runner := helper.TestRunner(t, map[string]string{filename: tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues)
		})
	}
}
