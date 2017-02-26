package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterDuplicateIDDetector struct {
	*Detector
	IssueType     string
	Target        string
	DeepCheck     bool
	cacheClusters map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterDuplicateIDDetector() *AwsElastiCacheClusterDuplicateIDDetector {
	return &AwsElastiCacheClusterDuplicateIDDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
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

func (d *AwsElastiCacheClusterDuplicateIDDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	idToken, err := hclLiteralToken(item, "cluster_id")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	id, err := d.evalToString(idToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.cacheClusters[id] && !d.State.Exists(d.Target, hclObjectKeyText(item)) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is duplicate Cluster ID. It must be unique.", id),
			Line:    idToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
