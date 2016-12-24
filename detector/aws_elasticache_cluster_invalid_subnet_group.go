package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterInvalidSubnetGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidSubnetGroupDetector() *AwsElastiCacheClusterInvalidSubnetGroupDetector {
	return &AwsElastiCacheClusterInvalidSubnetGroupDetector{d}
}

func (d *AwsElastiCacheClusterInvalidSubnetGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_elasticache_cluster") {
		return
	}

	validCacheSubnetGroups := map[string]bool{}
	if d.ResponseCache.DescribeCacheSubnetGroupsOutput == nil {
		resp, err := d.AwsClient.Elasticache.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeCacheSubnetGroupsOutput = resp
	}
	for _, subnetGroup := range d.ResponseCache.DescribeCacheSubnetGroupsOutput.CacheSubnetGroups {
		validCacheSubnetGroups[*subnetGroup.CacheSubnetGroupName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elasticache_cluster").Items {
			subnetGroupToken, err := hclLiteralToken(item, "subnet_group_name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			subnetGroup, err := d.evalToString(subnetGroupToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validCacheSubnetGroups[subnetGroup] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid subnet group name.", subnetGroup),
					Line:    subnetGroupToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
