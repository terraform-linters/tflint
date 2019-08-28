package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsElastiCacheClusterInvalidSubnetGroupRule checks whether subnet groups actually exists
type AwsElastiCacheClusterInvalidSubnetGroupRule struct {
	resourceType      string
	attributeName     string
	cacheSubnetGroups map[string]bool
	dataPrepared      bool
}

// NewAwsElastiCacheClusterInvalidSubnetGroupRule returns new rule with default attributes
func NewAwsElastiCacheClusterInvalidSubnetGroupRule() *AwsElastiCacheClusterInvalidSubnetGroupRule {
	return &AwsElastiCacheClusterInvalidSubnetGroupRule{
		resourceType:      "aws_elasticache_cluster",
		attributeName:     "subnet_group_name",
		cacheSubnetGroups: map[string]bool{},
		dataPrepared:      false,
	}
}

// Name returns the rule name
func (r *AwsElastiCacheClusterInvalidSubnetGroupRule) Name() string {
	return "aws_elasticache_cluster_invalid_subnet_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsElastiCacheClusterInvalidSubnetGroupRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsElastiCacheClusterInvalidSubnetGroupRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsElastiCacheClusterInvalidSubnetGroupRule) Link() string {
	return ""
}

// Check checks whether `subnet_group_name` are included in the list retrieved by `DescribeCacheSubnetGroups`
func (r *AwsElastiCacheClusterInvalidSubnetGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch cache subnet groups")
			resp, err := runner.AwsClient.ElastiCache.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing cache subnet groups",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, subnetGroup := range resp.CacheSubnetGroups {
				r.cacheSubnetGroups[*subnetGroup.CacheSubnetGroupName] = true
			}
			r.dataPrepared = true
		}

		var subnetGroup string
		err := runner.EvaluateExpr(attribute.Expr, &subnetGroup)

		return runner.EnsureNoError(err, func() error {
			if !r.cacheSubnetGroups[subnetGroup] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid subnet group name.", subnetGroup),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
