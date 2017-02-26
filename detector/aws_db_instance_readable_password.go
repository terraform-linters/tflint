package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

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

func (d *AwsDBInstanceReadablePasswordDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	passwordToken, err := hclLiteralToken(item, "password")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	_, err = d.evalToString(passwordToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	issue := &issue.Issue{
		Type:    d.IssueType,
		Message: "Password for the master DB user is readable. recommend using environment variables.",
		Line:    passwordToken.Pos.Line,
		File:    file,
	}
	*issues = append(*issues, issue)
}
