package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterPreviousTypeDetector struct {
	*Detector
}

func (d *Detector) CreateAwsElastiCacheClusterPreviousTypeDetector() *AwsElastiCacheClusterPreviousTypeDetector {
	return &AwsElastiCacheClusterPreviousTypeDetector{d}
}

func (d *AwsElastiCacheClusterPreviousTypeDetector) Detect(issues *[]*issue.Issue) {
	var previousNodeType = map[string]bool{
		"cache.m1.small":   true,
		"cache.m1.medium":  true,
		"cache.m1.large":   true,
		"cache.m1.xlarge":  true,
		"cache.m2.xlarge":  true,
		"cache.m2.2xlarge": true,
		"cache.m2.4xlarge": true,
		"cache.c1.xlarge":  true,
		"cache.t1.micro":   true,
	}

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

			if previousNodeType[nodeType] {
				issue := &issue.Issue{
					Type:    "WARNING",
					Message: fmt.Sprintf("\"%s\" is previous generation node type.", nodeType),
					Line:    nodeTypeToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
