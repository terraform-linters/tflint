package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterPreviousTypeDetector struct {
	*Detector
	previousNodeTypes map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterPreviousTypeDetector() *AwsElastiCacheClusterPreviousTypeDetector {
	nd := &AwsElastiCacheClusterPreviousTypeDetector{
		Detector:          d,
		previousNodeTypes: map[string]bool{},
	}
	nd.Name = "aws_elasticache_cluster_previous_type"
	nd.IssueType = issue.WARNING
	nd.TargetType = "resource"
	nd.Target = "aws_elasticache_cluster"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_elasticache_cluster_previous_type.md"
	nd.Enabled = true
	return nd
}

func (d *AwsElastiCacheClusterPreviousTypeDetector) PreProcess() {
	d.previousNodeTypes = map[string]bool{
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
}

func (d *AwsElastiCacheClusterPreviousTypeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	nodeTypeToken, ok := resource.GetToken("node_type")
	if !ok {
		return
	}
	nodeType, err := d.evalToString(nodeTypeToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.previousNodeTypes[nodeType] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is previous generation node type.", nodeType),
			Line:     nodeTypeToken.Pos.Line,
			File:     nodeTypeToken.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}
