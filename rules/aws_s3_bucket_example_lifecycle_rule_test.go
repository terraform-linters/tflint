package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_AwsS3BucketExampleLifecycleRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "issue found",
			Content: `
resource "aws_s3_bucket" "bucket" {
  lifecycle_rule {
    enabled = false

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }
  }
}`,
			Expected: helper.Issues{
				{
					Rule:    NewAwsS3BucketExampleLifecycleRule(),
					Message: "`lifecycle_rule` block found",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 17},
					},
				},
				{
					Rule:    NewAwsS3BucketExampleLifecycleRule(),
					Message: "`enabled` attribute found",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 15},
						End:      hcl.Pos{Line: 4, Column: 20},
					},
				},
				{
					Rule:    NewAwsS3BucketExampleLifecycleRule(),
					Message: "`transition` block found",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 6, Column: 5},
						End:      hcl.Pos{Line: 6, Column: 15},
					},
				},
			},
		},
	}

	rule := NewAwsS3BucketExampleLifecycleRule()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"resource.tf": test.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
