package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterInvalidParameterGroupDetector struct {
	*Detector
	IssueType            string
	TargetType           string
	Target               string
	DeepCheck            bool
	cacheParameterGroups map[string]bool
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidParameterGroupDetector() *AwsElastiCacheClusterInvalidParameterGroupDetector {
	return &AwsElastiCacheClusterInvalidParameterGroupDetector{
		Detector:             d,
		IssueType:            issue.ERROR,
		TargetType:           "resource",
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

func (d *AwsElastiCacheClusterInvalidParameterGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	parameterGroupToken, ok := resource.GetToken("parameter_group_name")
	if !ok {
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
			File:    parameterGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
