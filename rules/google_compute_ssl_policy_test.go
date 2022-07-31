package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_GoogleComputeSSLPolicyRule(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Config   string
		Expected helper.Issues
	}{
		{
			Name: "issue found",
			Content: `
resource "google_compute_ssl_policy" "allowed" {
  min_tls_version = "TLS_1_1"
}`,
			Config: `
rule "google_compute_ssl_policy" {
  enabled          = true
  allowed_versions = ["TLS_1_2"]
}`,
			Expected: helper.Issues{
				{
					Rule:    NewGoogleComputeSSLPolicyRule(),
					Message: `"TLS_1_1" is not allowed`,
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 21},
						End:      hcl.Pos{Line: 3, Column: 30},
					},
				},
			},
		},
		{
			Name: "issue not found",
			Content: `
resource "google_compute_ssl_policy" "allowed" {
  min_tls_version = "TLS_1_1"
}`,
			Config: `
rule "google_compute_ssl_policy" {
  enabled          = true
  allowed_versions = ["TLS_1_1", "TLS_1_2"]
}`,
			Expected: helper.Issues{},
		},
	}

	rule := NewGoogleComputeSSLPolicyRule()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"resource.tf": test.Content, ".tflint.hcl": test.Config})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, test.Expected, runner.Issues)
		})
	}
}
