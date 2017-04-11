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
		Name              string
		Src               string
		State             string
		SecurityGroups    []*ec2.SecurityGroup
		AccountAttributes []*ec2.AccountAttribute
		Issues            []*issue.Issue
	}{
		{
			Name: "security group name is duplicate",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
    vpc_id = "vpc-1234abcd"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("vpc-1234abcd"),
						},
					},
				},
			},
			Issues: []*issue.Issue{
				{
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
				{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("vpc-1234abcd"),
						},
					},
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
				{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-abcd1234"),
				},
				{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-abcd1234"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("vpc-1234abcd"),
						},
					},
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
				{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("vpc-abcd1234"),
						},
					},
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
				{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("vpc-1234abcd"),
						},
					},
				},
			},
			Issues: []*issue.Issue{
				{
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
				{
					GroupName: aws.String("default"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
				{
					GroupName: aws.String("custom"),
					VpcId:     aws.String("vpc-1234abcd"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("vpc-1234abcd"),
						},
					},
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "security group name is duplicate when omitted vpc_id on EC2-Classic",
			Src: `
resource "aws_security_group" "test" {
    name   = "default"
}`,
			SecurityGroups: []*ec2.SecurityGroup{
				{
					GroupName: aws.String("default"),
				},
				{
					GroupName: aws.String("custom"),
				},
			},
			AccountAttributes: []*ec2.AccountAttribute{
				{
					AttributeName: aws.String("supported-platforms"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("EC2"),
						},
						{
							AttributeValue: aws.String("VPC"),
						},
					},
				},
				{
					AttributeName: aws.String("default-vpc"),
					AttributeValues: []*ec2.AccountAttributeValue{
						{
							AttributeValue: aws.String("none"),
						},
					},
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"default\" is duplicate name. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
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
		ec2mock.EXPECT().DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{}).Return(&ec2.DescribeAccountAttributesOutput{
			AccountAttributes: tc.AccountAttributes,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsSecurityGroupDuplicateDetector",
			tc.Src,
			tc.State,
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
