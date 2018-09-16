package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteSpecifiedMultipleTargetsRule checks whether a route definition has multiple routing targets
type AwsRouteSpecifiedMultipleTargetsRule struct {
	resourceType string
}

// NewAwsRouteSpecifiedMultipleTargetsRule returns new rule with default attributes
func NewAwsRouteSpecifiedMultipleTargetsRule() *AwsRouteSpecifiedMultipleTargetsRule {
	return &AwsRouteSpecifiedMultipleTargetsRule{
		resourceType: "aws_route",
	}
}

// Name returns the rule name
func (r *AwsRouteSpecifiedMultipleTargetsRule) Name() string {
	return "aws_route_specified_multiple_targets"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteSpecifiedMultipleTargetsRule) Enabled() bool {
	return true
}

// Check checks whether a resource defines `gateway_id`, `egress_only_gateway_id`, `nat_gateway_id`
// `instance_id`, `vpc_peering_connection_id` or `network_interface_id` at the same time
func (r *AwsRouteSpecifiedMultipleTargetsRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, resource := range runner.LookupResourcesByType(r.resourceType) {
		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name: "gateway_id",
				},
				{
					Name: "egress_only_gateway_id",
				},
				{
					Name: "nat_gateway_id",
				},
				{
					Name: "instance_id",
				},
				{
					Name: "vpc_peering_connection_id",
				},
				{
					Name: "network_interface_id",
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		if len(body.Attributes) > 1 {
			runner.Issues = append(runner.Issues, &issue.Issue{
				Detector: r.Name(),
				Type:     issue.ERROR,
				Message:  "More than one routing target specified. It must be one.",
				Line:     resource.DeclRange.Start.Line,
				File:     runner.GetFileName(resource.DeclRange.Filename),
				Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_route_specified_multiple_targets.md",
			})
		}
	}

	return nil
}
