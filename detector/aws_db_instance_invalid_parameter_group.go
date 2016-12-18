package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceInvalidParameterGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceInvalidParameterGroupDetector() *AwsDBInstanceInvalidParameterGroupDetector {
	return &AwsDBInstanceInvalidParameterGroupDetector{d}
}

func (d *AwsDBInstanceInvalidParameterGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_db_instance") {
		return
	}

	validDBParameterGroups := map[string]bool{}
	if d.ResponseCache.DescribeDBParameterGroupsOutput == nil {
		resp, err := d.AwsClient.Rds.DescribeDBParameterGroups(&rds.DescribeDBParameterGroupsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeDBParameterGroupsOutput = resp
	}
	for _, parameterGroup := range d.ResponseCache.DescribeDBParameterGroupsOutput.DBParameterGroups {
		validDBParameterGroups[*parameterGroup.DBParameterGroupName] = true
	}

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

			if !validDBParameterGroups[parameterGroup] {
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
