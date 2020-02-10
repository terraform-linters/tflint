package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsElasticBeanstalkApplicationTagsRule checks whether the resource is tagged correctly
type AwsElasticBeanstalkApplicationTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsElasticBeanstalkApplicationTagsRule returns new tags rule with default attributes
func NewAwsElasticBeanstalkApplicationTagsRule() *AwsElasticBeanstalkApplicationTagsRule {
	return &AwsElasticBeanstalkApplicationTagsRule{
		resourceType:  "aws_elastic_beanstalk_application",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsElasticBeanstalkApplicationTagsRule) Name() string {
	return "aws_resource_tags_aws_elastic_beanstalk_application"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsElasticBeanstalkApplicationTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsElasticBeanstalkApplicationTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsElasticBeanstalkApplicationTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsElasticBeanstalkApplicationTagsRule) Check(runner *tflint.Runner) error {
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
