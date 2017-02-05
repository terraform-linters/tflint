package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsDBInstanceInvalidParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*rds.DBParameterGroup
		Issues   []*issue.Issue
	}{
		{
			Name: "parameter_group_name is invalid",
			Src: `
resource "aws_db_instance" "mysql" {
    parameter_group_name = "app-server"
}`,
			Response: []*rds.DBParameterGroup{
				&rds.DBParameterGroup{
					DBParameterGroupName: aws.String("app-server1"),
				},
				&rds.DBParameterGroup{
					DBParameterGroupName: aws.String("app-server2"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"app-server\" is invalid parameter group name.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "parameter_group_name is valid",
			Src: `
resource "aws_db_instance" "mysql" {
    parameter_group_name = "app-server"
}`,
			Response: []*rds.DBParameterGroup{
				&rds.DBParameterGroup{
					DBParameterGroupName: aws.String("app-server1"),
				},
				&rds.DBParameterGroup{
					DBParameterGroupName: aws.String("app-server2"),
				},
				&rds.DBParameterGroup{
					DBParameterGroupName: aws.String("app-server"),
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
		rdsmock.EXPECT().DescribeDBParameterGroups(&rds.DescribeDBParameterGroupsInput{}).Return(&rds.DescribeDBParameterGroupsOutput{
			DBParameterGroups: tc.Response,
		}, nil)
		awsClient.Rds = rdsmock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsDBInstanceInvalidParameterGroupDetector",
			tc.Src,
			"",
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
