package detector

import (
	"fmt"
	"regexp"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsElastiCacheClusterDefaultParameterGroupDetector struct {
	*Detector
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
}

func (d *Detector) CreateAwsElastiCacheClusterDefaultParameterGroupDetector() *AwsElastiCacheClusterDefaultParameterGroupDetector {
	return &AwsElastiCacheClusterDefaultParameterGroupDetector{
		Detector:   d,
		IssueType:  issue.NOTICE,
		TargetType: "resource",
		Target:     "aws_elasticache_cluster",
		DeepCheck:  false,
	}
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
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is default parameter group. You cannot edit it.", parameterGroup),
			Line:    parameterGroupToken.Pos.Line,
			File:    parameterGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}

func (d *AwsElastiCacheClusterDefaultParameterGroupDetector) isDefaultCacheParameterGroup(s string) bool {
	return regexp.MustCompile("^default").Match([]byte(s))
}
