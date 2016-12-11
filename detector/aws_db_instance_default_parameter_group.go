package detector

import (
	"fmt"
	"regexp"

	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceDefaultParameterGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceDefaultParameterGroupDetector() *AwsDBInstanceDefaultParameterGroupDetector {
	return &AwsDBInstanceDefaultParameterGroupDetector{d}
}

func (d *AwsDBInstanceDefaultParameterGroupDetector) Detect(issues *[]*issue.Issue) {
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

			if d.isDefaultDbParameterGroup(parameterGroup) {
				issue := &issue.Issue{
					Type:    "NOTICE",
					Message: fmt.Sprintf("\"%s\" is default parameter group. You cannot edit it.", parameterGroup),
					Line:    parameterGroupToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}

func (d *AwsDBInstanceDefaultParameterGroupDetector) isDefaultDbParameterGroup(s string) bool {
	return regexp.MustCompile("^default").Match([]byte(s))
}
