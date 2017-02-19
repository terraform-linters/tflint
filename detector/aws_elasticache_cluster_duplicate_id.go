package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterDuplicateIDDetector struct {
	*Detector
}

func (d *Detector) CreateAwsElastiCacheClusterDuplicateIDDetector() *AwsElastiCacheClusterDuplicateIDDetector {
	return &AwsElastiCacheClusterDuplicateIDDetector{d}
}

func (d *AwsElastiCacheClusterDuplicateIDDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_elasticache_cluster") {
		return
	}

	existCacheClusterId := map[string]bool{}
	resp, err := d.AwsClient.DescribeCacheClusters()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	for _, cacheCluster := range resp.CacheClusters {
		existCacheClusterId[*cacheCluster.CacheClusterId] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elasticache_cluster").Items {
			idToken, err := hclLiteralToken(item, "cluster_id")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			id, err := d.evalToString(idToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if existCacheClusterId[id] && !d.State.Exists("aws_elasticache_cluster", hclObjectKeyText(item)) {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is duplicate Cluster ID. It must be unique.", id),
					Line:    idToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
