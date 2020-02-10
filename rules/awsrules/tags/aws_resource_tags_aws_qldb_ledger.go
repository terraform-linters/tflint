package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsQlDBLedgerTagsRule checks whether the resource is tagged correctly
type AwsQlDBLedgerTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsQlDBLedgerTagsRule returns new tags rule with default attributes
func NewAwsQlDBLedgerTagsRule() *AwsQlDBLedgerTagsRule {
	return &AwsQlDBLedgerTagsRule{
		resourceType:  "aws_qldb_ledger",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsQlDBLedgerTagsRule) Name() string {
	return "aws_resource_tags_aws_qldb_ledger"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsQlDBLedgerTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsQlDBLedgerTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsQlDBLedgerTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsQlDBLedgerTagsRule) Check(runner *tflint.Runner) error {
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var resourceTags map[string]string
		err := runner.EvaluateExpr(attribute.Expr, &resourceTags)
		tags := []string{}
		for k, _ := range resourceTags {
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
				wanted := strings.Join(sort.StringSlice(configTags), ",")
				found := strings.Join(sort.StringSlice(tags), ",")
				runner.EmitIssue(r, fmt.Sprintf("Wanted tags: %v, found: %v\n", wanted, found), attribute.Expr.Range())
			}
			return nil
		})
	})
}
