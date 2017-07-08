package detector

import (
	"fmt"
	"regexp"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceDefaultParameterGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceDefaultParameterGroupDetector() *AwsDBInstanceDefaultParameterGroupDetector {
	nd := &AwsDBInstanceDefaultParameterGroupDetector{Detector: d}
	nd.Name = "aws_db_instance_default_parameter_group"
	nd.IssueType = issue.NOTICE
	nd.TargetType = "resource"
	nd.Target = "aws_db_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_default_parameter_group.md"
	return nd
}

func (d *AwsDBInstanceDefaultParameterGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	parameterGroupToken, ok := resource.GetToken("parameter_group_name")
	if !ok {
		return
	}
	parameterGroup, err := d.evalToString(parameterGroupToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.isDefaultDbParameterGroup(parameterGroup) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is default parameter group. You cannot edit it.", parameterGroup),
			Line:    parameterGroupToken.Pos.Line,
			File:    parameterGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}

func (d *AwsDBInstanceDefaultParameterGroupDetector) isDefaultDbParameterGroup(s string) bool {
	return regexp.MustCompile("^default").Match([]byte(s))
}
