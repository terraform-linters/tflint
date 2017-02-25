package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceInvalidOptionGroupDetector struct {
	*Detector
	IssueType    string
	Target       string
	DeepCheck    bool
	optionGroups map[string]bool
}

func (d *Detector) CreateAwsDBInstanceInvalidOptionGroupDetector() *AwsDBInstanceInvalidOptionGroupDetector {
	return &AwsDBInstanceInvalidOptionGroupDetector{
		Detector:     d,
		IssueType:    issue.ERROR,
		Target:       "aws_db_instance",
		DeepCheck:    true,
		optionGroups: map[string]bool{},
	}
}

func (d *AwsDBInstanceInvalidOptionGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeOptionGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, optionGroup := range resp.OptionGroupsList {
		d.optionGroups[*optionGroup.OptionGroupName] = true
	}
}

func (d *AwsDBInstanceInvalidOptionGroupDetector) Detect(issues *[]*issue.Issue) {
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

			if !d.optionGroups[optionGroup] {
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
