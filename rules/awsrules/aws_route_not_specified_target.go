package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsRouteNotSpecifiedTargetRule checks whether a route definition has a routing target
type AwsRouteNotSpecifiedTargetRule struct {
	resourceType string
}

// NewAwsRouteNotSpecifiedTargetRule returns new rule with default attributes
func NewAwsRouteNotSpecifiedTargetRule() *AwsRouteNotSpecifiedTargetRule {
	return &AwsRouteNotSpecifiedTargetRule{
		resourceType: "aws_route",
	}
}

// Name returns the rule name
func (r *AwsRouteNotSpecifiedTargetRule) Name() string {
	return "aws_route_not_specified_target"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRouteNotSpecifiedTargetRule) Enabled() bool {
	return true
}

// Check checks whether `gateway_id`, `egress_only_gateway_id`, `nat_gateway_id`, `instance_id`
// `vpc_peering_connection_id` or `network_interface_id` is defined in a resource
func (r *AwsRouteNotSpecifiedTargetRule) Check(runner *tflint.Runner) error {
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

		if len(body.Attributes) == 0 {
			runner.Issues = append(runner.Issues, &issue.Issue{
				Detector: r.Name(),
				Type:     issue.ERROR,
				Message:  "The routing target is not specified, each routing must contain either a gateway_id, egress_only_gateway_id a nat_gateway_id, an instance_id or a vpc_peering_connection_id or a network_interface_id.",
				Line:     resource.DeclRange.Start.Line,
				File:     runner.GetFileName(resource.DeclRange.Filename),
				Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_route_not_specified_target.md",
			})
		}
	}

	return nil
}
