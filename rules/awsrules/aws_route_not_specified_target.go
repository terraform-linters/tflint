package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/project"
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

// Type returns the rule severity
func (r *AwsRouteNotSpecifiedTargetRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsRouteNotSpecifiedTargetRule) Link() string {
	return project.ReferenceLink(r.Name())
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
				{
					Name: "transit_gateway_id",
				},
			},
		})
		if diags.HasErrors() {
			return diags
		}

		var nullAttributes int
		for _, attribute := range body.Attributes {
			if runner.IsNullExpr(attribute.Expr) {
				nullAttributes = nullAttributes + 1
			}
		}

		if len(body.Attributes)-nullAttributes == 0 {
			runner.EmitIssue(
				r,
				"The routing target is not specified, each aws_route must contain either egress_only_gateway_id, gateway_id, instance_id, nat_gateway_id, network_interface_id, transit_gateway_id, or vpc_peering_connection_id.",
				resource.DeclRange,
			)
		}
	}

	return nil
}
