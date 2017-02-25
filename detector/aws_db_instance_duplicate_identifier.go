package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceDuplicateIdentifierDetector struct {
	*Detector
	IssueType   string
	Target      string
	DeepCheck   bool
	identifiers map[string]bool
}

func (d *Detector) CreateAwsDBInstanceDuplicateIdentifierDetector() *AwsDBInstanceDuplicateIdentifierDetector {
	return &AwsDBInstanceDuplicateIdentifierDetector{
		Detector:    d,
		IssueType:   issue.ERROR,
		Target:      "aws_db_instance",
		DeepCheck:   true,
		identifiers: map[string]bool{},
	}
}

func (d *AwsDBInstanceDuplicateIdentifierDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeDBInstances()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, dbInstance := range resp.DBInstances {
		d.identifiers[*dbInstance.DBInstanceIdentifier] = true
	}
}

func (d *AwsDBInstanceDuplicateIdentifierDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
			identifierToken, err := hclLiteralToken(item, "identifier")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			identifier, err := d.evalToString(identifierToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if d.identifiers[identifier] && !d.State.Exists("aws_db_instance", hclObjectKeyText(item)) {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is duplicate identifier. It must be unique.", identifier),
					Line:    identifierToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
