package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterPreviousTypeDetector struct {
	*Detector
	IssueType         string
	Target            string
	DeepCheck         bool
	previousNodeTypes map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterPreviousTypeDetector() *AwsElastiCacheClusterPreviousTypeDetector {
	return &AwsElastiCacheClusterPreviousTypeDetector{
		Detector:          d,
		IssueType:         issue.WARNING,
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

func (d *AwsElastiCacheClusterPreviousTypeDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	nodeTypeToken, err := hclLiteralToken(item, "node_type")
	if err != nil {
		d.Logger.Error(err)
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
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
