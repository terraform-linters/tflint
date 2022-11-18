package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TestingAssertionsExampleRule checks whether ...
type TestingAssertionsExampleRule struct {
	tflint.DefaultRule
}

// NewTestingAssertionsExampleRule returns a new rule
func NewTestingAssertionsExampleRule() *TestingAssertionsExampleRule {
	return &TestingAssertionsExampleRule{}
}

// Name returns the rule name
func (r *TestingAssertionsExampleRule) Name() string {
	return "testing_assertions_example"
}

// Enabled returns whether the rule is enabled by default
func (r *TestingAssertionsExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TestingAssertionsExampleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TestingAssertionsExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *TestingAssertionsExampleRule) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("testing_assertions", &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "equal",
				LabelNames: []string{"name"},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		for _, equal := range resource.Body.Blocks {
			if err := runner.EmitIssue(r, fmt.Sprintf("equal block found: label=%s", equal.Labels[0]), equal.DefRange); err != nil {
				return err
			}
		}
	}

	return nil
}
