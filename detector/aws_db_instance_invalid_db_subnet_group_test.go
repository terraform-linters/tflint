package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsDBInstanceInvalidDBSubnetGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*rds.DBSubnetGroup
		Issues   []*issue.Issue
	}{
		{
			Name: "db_subnet_group_name is invalid",
			Src: `
resource "aws_db_instance" "mysql" {
    db_subnet_group_name = "app-server"
}`,
			Response: []*rds.DBSubnetGroup{
				{
					DBSubnetGroupName: aws.String("app-server1"),
				},
				{
					DBSubnetGroupName: aws.String("app-server2"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"app-server\" is invalid DB subnet group name.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "db_subnet_group_name is valid",
			Src: `
resource "aws_db_instance" "mysql" {
    db_subnet_group_name = "app-server"
}`,
			Response: []*rds.DBSubnetGroup{
				{
					DBSubnetGroupName: aws.String("app-server1"),
				},
				{
					DBSubnetGroupName: aws.String("app-server2"),
				},
				{
					DBSubnetGroupName: aws.String("app-server"),
				},
			},
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		c := config.Init()
		c.DeepCheck = true

		awsClient := c.NewAwsClient()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		rdsmock := mock.NewMockRDSAPI(ctrl)
		rdsmock.EXPECT().DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{}).Return(&rds.DescribeDBSubnetGroupsOutput{
			DBSubnetGroups: tc.Response,
		}, nil)
		awsClient.Rds = rdsmock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsDBInstanceInvalidDBSubnetGroupDetector",
			tc.Src,
			"",
			c,
			awsClient,
			&issues,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
