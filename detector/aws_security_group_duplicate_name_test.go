package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsSecurityGroupDuplicateName(t *testing.T) {
	cases := []struct {
		Name           string
		Src            string
		State          string
		SecurityGroups []*ec2.SecurityGroup
		Vpcs           []*ec2.Vpc
		Issues         []*issue.Issue
	}{
		{
			Name: "security group name is duplicate",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
    vpc_id = "vpc-1234abcd"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				&ec2.SecurityGroup{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"default\" is duplicate name. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "security group name is unique",
			Src: `
resource "aws_security_group" "test" {
    name   = "latest"
    vpc_id = "vpc-1234abcd"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				&ec2.SecurityGroup{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "security group name is duplicate, but in the another VPC",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
    vpc_id = "vpc-1234abcd"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-abcd1234"),
				},
				&ec2.SecurityGroup{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-abcd1234"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "security group name is duplicate in default vpc when omitted vpc_id",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				&ec2.SecurityGroup{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			Vpcs: []*ec2.Vpc{
				&ec2.Vpc{
					VpcId:     aws.String("vpc-1234abcd"),
					IsDefault: aws.Bool(false),
				},
				&ec2.Vpc{
					VpcId:     aws.String("vpc-abcd1234"),
					IsDefault: aws.Bool(true),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "security group name is duplicate in another vpc when omitted vpc_id",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				&ec2.SecurityGroup{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			Vpcs: []*ec2.Vpc{
				&ec2.Vpc{
					VpcId:     aws.String("vpc-1234abcd"),
					IsDefault: aws.Bool(true),
				},
				&ec2.Vpc{
					VpcId:     aws.String("vpc-abcd1234"),
					IsDefault: aws.Bool(false),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"default\" is duplicate name. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "security group name is duplicate, but exists in state",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
    vpc_id = "vpc-1234abcd"
}`,
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_security_group.test": {
                    "type": "aws_security_group",
                    "depends_on": [],
                    "primary": {
                        "id": "sg-1234abcd",
                        "attributes": {
                            "id": "sg-1234abcd",
                            "name": "default",
                            "owner_id": "123456789",
                            "vpc_id": "vpc-1234abcd"
                        }
                    }
                }
            }
        }
    ]
}
`,
			SecurityGroups: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				&ec2.SecurityGroup{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
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
		ec2mock := mock.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.SecurityGroups,
		}, nil)
		if tc.Vpcs != nil {
			ec2mock.EXPECT().DescribeVpcs(&ec2.DescribeVpcsInput{}).Return(&ec2.DescribeVpcsOutput{
				Vpcs: tc.Vpcs,
			}, nil)
		}
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsSecurityGroupDuplicateDetector",
			tc.Src,
			tc.State,
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
