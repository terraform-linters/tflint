package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterDuplicateIDDetector struct {
	*Detector
	IssueType     string
	TargetType    string
	Target        string
	DeepCheck     bool
	cacheClusters map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterDuplicateIDDetector() *AwsElastiCacheClusterDuplicateIDDetector {
	return &AwsElastiCacheClusterDuplicateIDDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
		TargetType:    "resource",
		Target:        "aws_elasticache_cluster",
		DeepCheck:     true,
		cacheClusters: map[string]bool{},
	}
}

func (d *AwsElastiCacheClusterDuplicateIDDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeCacheClusters()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, cacheCluster := range resp.CacheClusters {
		d.cacheClusters[*cacheCluster.CacheClusterId] = true
	}
}

func (d *AwsElastiCacheClusterDuplicateIDDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	idToken, ok := resource.GetToken("cluster_id")
	if !ok {
		return
	}
	id, err := d.evalToString(idToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.cacheClusters[id] && !d.State.Exists(d.Target, resource.Id) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is duplicate Cluster ID. It must be unique.", id),
			Line:    idToken.Pos.Line,
			File:    idToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
