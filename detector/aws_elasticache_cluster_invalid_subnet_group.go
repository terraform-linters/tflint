package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterInvalidSubnetGroupDetector struct {
	*Detector
	IssueType         string
	Target            string
	DeepCheck         bool
	cacheSubnetGroups map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidSubnetGroupDetector() *AwsElastiCacheClusterInvalidSubnetGroupDetector {
	return &AwsElastiCacheClusterInvalidSubnetGroupDetector{
		Detector:          d,
		IssueType:         issue.ERROR,
		Target:            "aws_elasticache_cluster",
		DeepCheck:         true,
		cacheSubnetGroups: map[string]bool{},
	}
}

func (d *AwsElastiCacheClusterInvalidSubnetGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeCacheSubnetGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, subnetGroup := range resp.CacheSubnetGroups {
		d.cacheSubnetGroups[*subnetGroup.CacheSubnetGroupName] = true
	}
}

func (d *AwsElastiCacheClusterInvalidSubnetGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	subnetGroupToken, ok := resource.GetToken("subnet_group_name")
	if !ok {
		return
	}
	subnetGroup, err := d.evalToString(subnetGroupToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.cacheSubnetGroups[subnetGroup] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid subnet group name.", subnetGroup),
			Line:    subnetGroupToken.Pos.Line,
			File:    subnetGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
