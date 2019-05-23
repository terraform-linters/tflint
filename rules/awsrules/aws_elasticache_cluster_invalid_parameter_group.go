package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsElastiCacheClusterInvalidParameterGroupRule checks whether cache parameter group actually exists
type AwsElastiCacheClusterInvalidParameterGroupRule struct {
	resourceType         string
	attributeName        string
	cacheParameterGroups map[string]bool
	dataPrepared         bool
}

// NewAwsElastiCacheClusterInvalidParameterGroupRule returns new rule with default attributes
func NewAwsElastiCacheClusterInvalidParameterGroupRule() *AwsElastiCacheClusterInvalidParameterGroupRule {
	return &AwsElastiCacheClusterInvalidParameterGroupRule{
		resourceType:         "aws_elasticache_cluster",
		attributeName:        "parameter_group_name",
		cacheParameterGroups: map[string]bool{},
		dataPrepared:         false,
	}
}

// Name returns the rule name
func (r *AwsElastiCacheClusterInvalidParameterGroupRule) Name() string {
	return "aws_elasticache_cluster_invalid_parameter_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsElastiCacheClusterInvalidParameterGroupRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsElastiCacheClusterInvalidParameterGroupRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsElastiCacheClusterInvalidParameterGroupRule) Link() string {
	return ""
}

// Check checks whether `parameter_group_name` are included in the list retrieved by `DescribeCacheParameterGroups`
func (r *AwsElastiCacheClusterInvalidParameterGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch cache parameter groups")
			resp, err := runner.AwsClient.ElastiCache.DescribeCacheParameterGroups(&elasticache.DescribeCacheParameterGroupsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing cache parameter groups",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, parameterGroup := range resp.CacheParameterGroups {
				r.cacheParameterGroups[*parameterGroup.CacheParameterGroupName] = true
			}
			r.dataPrepared = true
		}

		var parameterGroup string
		err := runner.EvaluateExpr(attribute.Expr, &parameterGroup)

		return runner.EnsureNoError(err, func() error {
			if !r.cacheParameterGroups[parameterGroup] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid parameter group name.", parameterGroup),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
