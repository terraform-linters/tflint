package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterInvalidTypeDetector struct {
	*Detector
	nodeTypes map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidTypeDetector() *AwsElastiCacheClusterInvalidTypeDetector {
	nd := &AwsElastiCacheClusterInvalidTypeDetector{
		Detector:  d,
		nodeTypes: map[string]bool{},
	}
	nd.Name = "aws_elasticache_cluster_invalid_type"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_elasticache_cluster"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_elasticache_cluster_invalid_type.md"
	nd.Enabled = true
	return nd
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

func (d *AwsElastiCacheClusterInvalidTypeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	nodeTypeToken, ok := resource.GetToken("node_type")
	if !ok {
		return
	}
	nodeType, err := d.evalToString(nodeTypeToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.nodeTypes[nodeType] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid node type.", nodeType),
			Line:     nodeTypeToken.Pos.Line,
			File:     nodeTypeToken.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}
