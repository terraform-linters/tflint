package detector

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsElastiCacheClusterDefaultParameterGroupDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateAwsElastiCacheClusterDefaultParameterGroupDetector() *AwsElastiCacheClusterDefaultParameterGroupDetector {
	return &AwsElastiCacheClusterDefaultParameterGroupDetector{
		Detector:  d,
		IssueType: issue.NOTICE,
		Target:    "aws_elasticache_cluster",
		DeepCheck: false,
	}
}

func (d *AwsElastiCacheClusterDefaultParameterGroupDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
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

	if d.isDefaultCacheParameterGroup(parameterGroup) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is default parameter group. You cannot edit it.", parameterGroup),
			Line:    parameterGroupToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}

func (d *AwsElastiCacheClusterDefaultParameterGroupDetector) isDefaultCacheParameterGroup(s string) bool {
	return regexp.MustCompile("^default").Match([]byte(s))
}
