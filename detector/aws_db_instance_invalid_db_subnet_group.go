package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/wata727/tflint/issue"
)

type AwsDBInstanceInvalidDBSubnetGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstanceInvalidDBSubnetGroupDetector() *AwsDBInstanceInvalidDBSubnetGroupDetector {
	return &AwsDBInstanceInvalidDBSubnetGroupDetector{d}
}

func (d *AwsDBInstanceInvalidDBSubnetGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_db_instance") {
		return
	}

	validDBSubnetGroups := map[string]bool{}
	if d.ResponseCache.DescribeDBSubnetGroupsOutput == nil {
		resp, err := d.AwsClient.Rds.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeDBSubnetGroupsOutput = resp
	}
	for _, subnetGroup := range d.ResponseCache.DescribeDBSubnetGroupsOutput.DBSubnetGroups {
		validDBSubnetGroups[*subnetGroup.DBSubnetGroupName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
			subnetGroupToken, err := hclLiteralToken(item, "db_subnet_group_name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			subnetGroup, err := d.evalToString(subnetGroupToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validDBSubnetGroups[subnetGroup] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid DB subnet group name.", subnetGroup),
					Line:    subnetGroupToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
