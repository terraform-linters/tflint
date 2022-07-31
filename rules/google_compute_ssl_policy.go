package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// GoogleComputeSSLPolicyRule checks whether ...
type GoogleComputeSSLPolicyRule struct {
	tflint.DefaultRule
}

// GoogleComputeSSLPolicyRuleConfig is a config of GoogleComputeSSLPolicyRule
type GoogleComputeSSLPolicyRuleConfig struct {
	AllowedVersions []string `hclext:"allowed_versions"`
}

// NewGoogleComputeSSLPolicyRule returns a new rule
func NewGoogleComputeSSLPolicyRule() *GoogleComputeSSLPolicyRule {
	return &GoogleComputeSSLPolicyRule{}
}

// Name returns the rule name
func (r *GoogleComputeSSLPolicyRule) Name() string {
	return "google_compute_ssl_policy"
}

// Enabled returns whether the rule is enabled by default
func (r *GoogleComputeSSLPolicyRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *GoogleComputeSSLPolicyRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *GoogleComputeSSLPolicyRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *GoogleComputeSSLPolicyRule) Check(runner tflint.Runner) error {
	// This rule is an example to use custom rule config.
	config := &GoogleComputeSSLPolicyRuleConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), config); err != nil {
		return err
	}

	resources, err := runner.GetResourceContent("google_compute_ssl_policy", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{
			{Name: "min_tls_version"},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		attribute, exists := resource.Body.Attributes["min_tls_version"]
		if !exists {
			continue
		}

		var version string
		err := runner.EvaluateExpr(attribute.Expr, &version, nil)

		err = runner.EnsureNoError(err, func() error {
			for _, allow := range config.AllowedVersions {
				if version == allow {
					return nil
				}
			}
			return runner.EmitIssue(
				r,
				fmt.Sprintf(`"%s" is not allowed`, version),
				attribute.Expr.Range(),
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
