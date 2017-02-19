package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceInvalidOptionGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceInvalidOptionGroupDetector() *AwsDBInstanceInvalidOptionGroupDetector {
	return &AwsDBInstanceInvalidOptionGroupDetector{d}
}

func (d *AwsDBInstanceInvalidOptionGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_db_instance") {
		return
	}

	validOptionGroups := map[string]bool{}
	resp, err := d.AwsClient.DescribeOptionGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	for _, optionGroup := range resp.OptionGroupsList {
		validOptionGroups[*optionGroup.OptionGroupName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
			optionGroupToken, err := hclLiteralToken(item, "option_group_name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			optionGroup, err := d.evalToString(optionGroupToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validOptionGroups[optionGroup] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid option group name.", optionGroup),
					Line:    optionGroupToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
