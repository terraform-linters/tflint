package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
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

func (d *AwsDBInstanceInvalidParameterGroupDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
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

			if !d.parameterGroups[parameterGroup] {
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
