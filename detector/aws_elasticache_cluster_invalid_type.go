package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterInvalidTypeDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	nodeTypes map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidTypeDetector() *AwsElastiCacheClusterInvalidTypeDetector {
	return &AwsElastiCacheClusterInvalidTypeDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_elasticache_cluster",
		DeepCheck: false,
		nodeTypes: map[string]bool{},
	}
}

func (d *AwsElastiCacheClusterInvalidTypeDetector) PreProcess() {
	d.nodeTypes = map[string]bool{
		"cache.t2.micro":    true,
		"cache.t2.small":    true,
		"cache.t2.medium":   true,
		"cache.m3.medium":   true,
		"cache.m3.large":    true,
		"cache.m3.xlarge":   true,
		"cache.m3.2xlarge":  true,
		"cache.m4.large":    true,
		"cache.m4.xlarge":   true,
		"cache.m4.2xlarge":  true,
		"cache.m4.4xlarge":  true,
		"cache.m4.10xlarge": true,
		"cache.r3.large":    true,
		"cache.r3.xlarge":   true,
		"cache.r3.2xlarge":  true,
		"cache.r3.4xlarge":  true,
		"cache.r3.8xlarge":  true,
		"cache.m1.small":    true,
		"cache.m1.medium":   true,
		"cache.m1.large":    true,
		"cache.m1.xlarge":   true,
		"cache.m2.xlarge":   true,
		"cache.m2.2xlarge":  true,
		"cache.m2.4xlarge":  true,
		"cache.c1.xlarge":   true,
		"cache.t1.micro":    true,
	}
}

func (d *AwsElastiCacheClusterInvalidTypeDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elasticache_cluster").Items {
			nodeTypeToken, err := hclLiteralToken(item, "node_type")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			nodeType, err := d.evalToString(nodeTypeToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !d.nodeTypes[nodeType] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid node type.", nodeType),
					Line:    nodeTypeToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
