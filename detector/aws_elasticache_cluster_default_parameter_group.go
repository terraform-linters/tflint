package detector

import (
	"fmt"
	"regexp"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterDefaultParameterGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsElastiCacheClusterDefaultParameterGroupDetector() *AwsElastiCacheClusterDefaultParameterGroupDetector {
	nd := &AwsElastiCacheClusterDefaultParameterGroupDetector{Detector: d}
	nd.Name = "aws_elasticache_cluster_default_parameter_group"
	nd.IssueType = issue.NOTICE
	nd.TargetType = "resource"
	nd.Target = "aws_elasticache_cluster"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_elasticache_cluster_default_parameter_group.md"
	return nd
}

func (d *AwsElastiCacheClusterDefaultParameterGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	parameterGroupToken, ok := resource.GetToken("parameter_group_name")
	if !ok {
		return
	}
	parameterGroup, err := d.evalToString(parameterGroupToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.isDefaultCacheParameterGroup(parameterGroup) {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is default parameter group. You cannot edit it.", parameterGroup),
			Line:     parameterGroupToken.Pos.Line,
			File:     parameterGroupToken.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}

func (d *AwsElastiCacheClusterDefaultParameterGroupDetector) isDefaultCacheParameterGroup(s string) bool {
	return regexp.MustCompile("^default").Match([]byte(s))
}
