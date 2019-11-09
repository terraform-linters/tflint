package api

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/client"
	"github.com/terraform-linters/tflint/tflint"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func Test_APIListData(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected tflint.Issues
	}{
		{
			Name: "security group is invalid",
			Content: `
resource "aws_alb" "balancer" {
    security_groups = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
			Response: []*ec2.SecurityGroup{
				{
					GroupId: aws.String("sg-12345678"),
				},
				{
					GroupId: aws.String("sg-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 22},
					},
				},
			},
		},
		{
			Name: "security group is valid",
			Content: `
resource "aws_alb" "balancer" {
    security_groups = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
			Response: []*ec2.SecurityGroup{
				{
					GroupId: aws.String("sg-1234abcd"),
				},
				{
					GroupId: aws.String("sg-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
		{
			Name: "use list variables",
			Content: `
variable "security_groups" {
    default = ["sg-1234abcd", "sg-abcd1234"]
}

resource "aws_alb" "balancer" {
    security_groups = "${var.security_groups}"
}`,
			Response: []*ec2.SecurityGroup{
				{
					GroupId: aws.String("sg-12345678"),
				},
				{
					GroupId: aws.String("sg-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 23},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
				},
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 23},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		rule := NewAwsALBInvalidSecurityGroupRule()
		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func Test_APIStringData(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*rds.DBSubnetGroup
		Expected tflint.Issues
	}{
		{
			Name: "db_subnet_group_name is invalid",
			Content: `
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidDBSubnetGroupRule(),
					Message: "\"app-server\" is invalid DB subnet group name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 40},
					},
				},
			},
		},
		{
			Name: "db_subnet_group_name is valid",
			Content: `
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
			Expected: tflint.Issues{},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		mock := client.NewMockRDSAPI(ctrl)
		mock.EXPECT().DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{}).Return(&rds.DescribeDBSubnetGroupsOutput{
			DBSubnetGroups: tc.Response,
		}, nil)
		runner.AwsClient.RDS = mock

		rule := NewAwsDBInstanceInvalidDBSubnetGroupRule()
		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func Test_API_error(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response error
		Error    tflint.Error
	}{
		{
			Name: "API error",
			Content: `
resource "aws_alb" "balancer" {
    security_groups = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
			Response: errors.New("MissingRegion: could not find region configuration"),
			Error: tflint.Error{
				Code:    tflint.ExternalAPIError,
				Level:   tflint.ErrorLevel,
				Message: "An error occurred while invoking DescribeSecurityGroups; MissingRegion: could not find region configuration",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rule := NewAwsALBInvalidSecurityGroupRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(nil, tc.Response)
		runner.AwsClient.EC2 = mock

		err := rule.Check(runner)
		tflint.AssertAppError(t, tc.Error, err)
	}
}
