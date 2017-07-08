package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceReadablePasswordDetector struct {
	*Detector
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
}

func (d *Detector) CreateAwsDBInstanceReadablePasswordDetector() *AwsDBInstanceReadablePasswordDetector {
	return &AwsDBInstanceReadablePasswordDetector{
		Detector:   d,
		IssueType:  issue.WARNING,
		TargetType: "resource",
		Target:     "aws_db_instance",
		DeepCheck:  false,
	}
}

func (d *AwsDBInstanceReadablePasswordDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	passwordToken, ok := resource.GetToken("password")
	if !ok {
		return
	}
	_, err := d.evalToString(passwordToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	issue := &issue.Issue{
		Type:    d.IssueType,
		Message: "Password for the master DB user is readable. recommend using environment variables.",
		Line:    passwordToken.Pos.Line,
		File:    passwordToken.Pos.Filename,
	}
	*issues = append(*issues, issue)
}
