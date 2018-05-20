package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceReadablePasswordDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceReadablePasswordDetector() *AwsDBInstanceReadablePasswordDetector {
	nd := &AwsDBInstanceReadablePasswordDetector{Detector: d}
	nd.Name = "aws_db_instance_readable_password"
	nd.IssueType = issue.WARNING
	nd.TargetType = "resource"
	nd.Target = "aws_db_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_readable_password.md"
	nd.Enabled = true
	return nd
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
		Detector: d.Name,
		Type:     d.IssueType,
		Message:  "Password for the master DB user is readable. recommend using environment variables.",
		Line:     passwordToken.Pos.Line,
		File:     passwordToken.Pos.Filename,
		Link:     d.Link,
	}
	*issues = append(*issues, issue)
}
