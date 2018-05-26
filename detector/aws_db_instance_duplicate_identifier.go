package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceDuplicateIdentifierDetector struct {
	*Detector
	identifiers map[string]bool
}

func (d *Detector) CreateAwsDBInstanceDuplicateIdentifierDetector() *AwsDBInstanceDuplicateIdentifierDetector {
	nd := &AwsDBInstanceDuplicateIdentifierDetector{
		Detector:    d,
		identifiers: map[string]bool{},
	}
	nd.Name = "aws_db_instance_duplicate_identifier"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_db_instance"
	nd.DeepCheck = true
	nd.Enabled = true
	return nd
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

func (d *AwsDBInstanceDuplicateIdentifierDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	identifierToken, ok := resource.GetToken("identifier")
	if !ok {
		return
	}
	identifier, err := d.evalToString(identifierToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	identityCheckFunc := func(attributes map[string]string) bool { return attributes["identifier"] == identifier }
	if d.identifiers[identifier] && !d.State.Exists(d.Target, resource.Id, identityCheckFunc) {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is duplicate identifier. It must be unique.", identifier),
			Line:     identifierToken.Pos.Line,
			File:     identifierToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
