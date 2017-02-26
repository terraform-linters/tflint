package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterInvalidParameterGroupDetector struct {
	*Detector
	IssueType            string
	Target               string
	DeepCheck            bool
	cacheParameterGroups map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidParameterGroupDetector() *AwsElastiCacheClusterInvalidParameterGroupDetector {
	return &AwsElastiCacheClusterInvalidParameterGroupDetector{
		Detector:             d,
		IssueType:            issue.ERROR,
		Target:               "aws_elasticache_cluster",
		DeepCheck:            true,
		cacheParameterGroups: map[string]bool{},
	}
}

func (d *AwsElastiCacheClusterInvalidParameterGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeCacheParameterGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, parameterGroup := range resp.CacheParameterGroups {
		d.cacheParameterGroups[*parameterGroup.CacheParameterGroupName] = true
	}
}

func (d *AwsElastiCacheClusterInvalidParameterGroupDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	parameterGroupToken, err := hclLiteralToken(item, "parameter_group_name")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	parameterGroup, err := d.evalToString(parameterGroupToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.cacheParameterGroups[parameterGroup] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid parameter group name.", parameterGroup),
			Line:    parameterGroupToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
