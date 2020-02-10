package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsDxHostedTransitVirtualInterfaceAccepterTagsRule checks whether the resource is tagged correctly
type AwsDxHostedTransitVirtualInterfaceAccepterTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsDxHostedTransitVirtualInterfaceAccepterTagsRule returns new tags rule with default attributes
func NewAwsDxHostedTransitVirtualInterfaceAccepterTagsRule() *AwsDxHostedTransitVirtualInterfaceAccepterTagsRule {
	return &AwsDxHostedTransitVirtualInterfaceAccepterTagsRule{
		resourceType:  "aws_dx_hosted_transit_virtual_interface_accepter",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsDxHostedTransitVirtualInterfaceAccepterTagsRule) Name() string {
	return "aws_resource_tags_aws_dx_hosted_transit_virtual_interface_accepter"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDxHostedTransitVirtualInterfaceAccepterTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDxHostedTransitVirtualInterfaceAccepterTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsDxHostedTransitVirtualInterfaceAccepterTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsDxHostedTransitVirtualInterfaceAccepterTagsRule) Check(runner *tflint.Runner) error {
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var resourceTags map[string]string
		err := runner.EvaluateExpr(attribute.Expr, &resourceTags)
		tags := []string{}
		for k := range resourceTags {
			tags = append(tags, k)
		}

		return runner.EnsureNoError(err, func() error {
			configTags := runner.GetConfigTags()
			hash := make(map[string]bool)
			for _, k := range tags {
				hash[k] = true
			}
			var found []string
			for _, tag := range configTags {
				if _, ok := hash[tag]; ok {
					found = append(found, tag)
				}
			}
			if len(found) != len(configTags) {
				sort.Strings(configTags)
				sort.Strings(tags)
				wanted := strings.Join(configTags, ",")
				found := strings.Join(tags, ",")
				runner.EmitIssue(r, fmt.Sprintf("Wanted tags: %v, found: %v\n", wanted, found), attribute.Expr.Range())
			}
			return nil
		})
	})
}
