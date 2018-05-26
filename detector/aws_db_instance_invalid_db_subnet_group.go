package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceInvalidDBSubnetGroupDetector struct {
	*Detector
	subnetGroups map[string]bool
}

func (d *Detector) CreateAwsDBInstanceInvalidDBSubnetGroupDetector() *AwsDBInstanceInvalidDBSubnetGroupDetector {
	nd := &AwsDBInstanceInvalidDBSubnetGroupDetector{
		Detector:     d,
		subnetGroups: map[string]bool{},
	}
	nd.Name = "aws_db_instance_invalid_db_subnet_group"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_db_instance"
	nd.DeepCheck = true
	nd.Enabled = true
	return nd
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
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid DB subnet group name.", subnetGroup),
			Line:     subnetGroupToken.Pos.Line,
			File:     subnetGroupToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
