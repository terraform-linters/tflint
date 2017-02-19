package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterInvalidParameterGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsElastiCacheClusterInvalidParameterGroupDetector() *AwsElastiCacheClusterInvalidParameterGroupDetector {
	return &AwsElastiCacheClusterInvalidParameterGroupDetector{d}
}

func (d *AwsElastiCacheClusterInvalidParameterGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_elasticache_cluster") {
		return
	}

	validCacheParameterGroups := map[string]bool{}
	resp, err := d.AwsClient.DescribeCacheParameterGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	for _, parameterGroup := range resp.CacheParameterGroups {
		validCacheParameterGroups[*parameterGroup.CacheParameterGroupName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elasticache_cluster").Items {
			parameterGroupToken, err := hclLiteralToken(item, "parameter_group_name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			parameterGroup, err := d.evalToString(parameterGroupToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validCacheParameterGroups[parameterGroup] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid parameter group name.", parameterGroup),
					Line:    parameterGroupToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
