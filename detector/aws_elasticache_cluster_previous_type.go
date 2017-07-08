package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterPreviousTypeDetector struct {
	*Detector
	IssueType         string
	TargetType        string
	Target            string
	DeepCheck         bool
	previousNodeTypes map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterPreviousTypeDetector() *AwsElastiCacheClusterPreviousTypeDetector {
	return &AwsElastiCacheClusterPreviousTypeDetector{
		Detector:          d,
		IssueType:         issue.WARNING,
		TargetType:        "resource",
		Target:            "aws_elasticache_cluster",
		DeepCheck:         false,
		previousNodeTypes: map[string]bool{},
	}
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
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is previous generation node type.", nodeType),
			Line:    nodeTypeToken.Pos.Line,
			File:    nodeTypeToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
