package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsPrometheusWorkspaceUnignorableLoggingConfiguration checks whether ...
type AwsPrometheusWorkspaceUnignorableLoggingConfiguration struct {
	tflint.DefaultRule
}

// NewAwsPrometheusWorkspaceUnignorableLoggingConfigurationRule returns a new rule
func NewAwsPrometheusWorkspaceUnignorableLoggingConfigurationRule() *AwsPrometheusWorkspaceUnignorableLoggingConfiguration {
	return &AwsPrometheusWorkspaceUnignorableLoggingConfiguration{}
}

// Name returns the rule name
func (r *AwsPrometheusWorkspaceUnignorableLoggingConfiguration) Name() string {
	return "aws_prometheus_workspace_unignorable_logging_configuration"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsPrometheusWorkspaceUnignorableLoggingConfiguration) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsPrometheusWorkspaceUnignorableLoggingConfiguration) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *AwsPrometheusWorkspaceUnignorableLoggingConfiguration) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsPrometheusWorkspaceUnignorableLoggingConfiguration) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_prometheus_workspace", &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "lifecycle",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{
							Name: "ignore_changes",
						},
					},
				},
			},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		for _, lifecycleBlock := range resource.Body.Blocks {
			if ignoreChangesAttr, exists := lifecycleBlock.Body.Attributes["ignore_changes"]; exists {
				exprs, diags := hcl.ExprList(ignoreChangesAttr.Expr)
				if diags.HasErrors() {
					// Skip if the ignore_changes is not a list
					continue
				}

				for _, expr := range exprs {
					keyword := hcl.ExprAsKeyword(expr)
					if keyword == "logging_configuration" {
						err := runner.EmitIssue(
							r,
							"ignore_changes should not include logging_configuration",
							ignoreChangesAttr.Range,
						)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
