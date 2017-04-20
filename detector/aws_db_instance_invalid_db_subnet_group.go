package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceInvalidDBSubnetGroupDetector struct {
	*Detector
	IssueType    string
	Target       string
	DeepCheck    bool
	subnetGroups map[string]bool
}

func (d *Detector) CreateAwsDBInstanceInvalidDBSubnetGroupDetector() *AwsDBInstanceInvalidDBSubnetGroupDetector {
	return &AwsDBInstanceInvalidDBSubnetGroupDetector{
		Detector:     d,
		IssueType:    issue.ERROR,
		Target:       "aws_db_instance",
		DeepCheck:    true,
		subnetGroups: map[string]bool{},
	}
}

func (d *AwsDBInstanceInvalidDBSubnetGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeDBSubnetGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, subnetGroup := range resp.DBSubnetGroups {
		d.subnetGroups[*subnetGroup.DBSubnetGroupName] = true
	}
}

func (d *AwsDBInstanceInvalidDBSubnetGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	subnetGroupToken, ok := resource.GetToken("db_subnet_group_name")
	if !ok {
		return
	}
	subnetGroup, err := d.evalToString(subnetGroupToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.subnetGroups[subnetGroup] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid DB subnet group name.", subnetGroup),
			Line:    subnetGroupToken.Pos.Line,
			File:    subnetGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
