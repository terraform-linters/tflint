package detector

import "github.com/wata727/tflint/issue"

type AwsDBInstanceReadablePasswordDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateAwsDBInstanceReadablePasswordDetector() *AwsDBInstanceReadablePasswordDetector {
	return &AwsDBInstanceReadablePasswordDetector{
		Detector:  d,
		IssueType: issue.WARNING,
		Target:    "aws_db_instance",
		DeepCheck: false,
	}
}

func (d *AwsDBInstanceReadablePasswordDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
			passwordToken, err := hclLiteralToken(item, "password")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			_, err = d.evalToString(passwordToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			issue := &issue.Issue{
				Type:    "WARNING",
				Message: "Password for the master DB user is readable. recommend using environment variables.",
				Line:    passwordToken.Pos.Line,
				File:    filename,
			}
			*issues = append(*issues, issue)
		}
	}
}
