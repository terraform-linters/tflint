package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceInvalidParameterGroupDetector struct {
	*Detector
	IssueType       string
	Target          string
	DeepCheck       bool
	parameterGroups map[string]bool
}

func (d *Detector) CreateAwsDBInstanceInvalidParameterGroupDetector() *AwsDBInstanceInvalidParameterGroupDetector {
	return &AwsDBInstanceInvalidParameterGroupDetector{
		Detector:        d,
		IssueType:       issue.ERROR,
		Target:          "aws_db_instance",
		DeepCheck:       true,
		parameterGroups: map[string]bool{},
	}
}

func (d *AwsDBInstanceInvalidParameterGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeDBParameterGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, parameterGroup := range resp.DBParameterGroups {
		d.parameterGroups[*parameterGroup.DBParameterGroupName] = true
	}
}

func (d *AwsDBInstanceInvalidParameterGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	parameterGroupToken, ok := resource.GetToken("parameter_group_name")
	if !ok {
		return
	}
	parameterGroup, err := d.evalToString(parameterGroupToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.parameterGroups[parameterGroup] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid parameter group name.", parameterGroup),
			Line:    parameterGroupToken.Pos.Line,
			File:    parameterGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
