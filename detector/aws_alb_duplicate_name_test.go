package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/golang/mock/gomock"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsALBDuplicateName(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		State    string
		Response []*elbv2.LoadBalancer
		Issues   []*issue.Issue
	}{
		{
			Name: "ALB name is duplicate",
			Src: `
resource "aws_alb" "test" {
    name = "test-alb-tf"
}`,
			Response: []*elbv2.LoadBalancer{
				&elbv2.LoadBalancer{
					LoadBalancerName: aws.String("test-alb-tf"),
				},
				&elbv2.LoadBalancer{
					LoadBalancerName: aws.String("production-alb-tf"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"test-alb-tf\" is duplicate name. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "ALB name is unique",
			Src: `
resource "aws_alb" "test" {
    name = "test-alb-tf"
}`,
			Response: []*elbv2.LoadBalancer{
				&elbv2.LoadBalancer{
					LoadBalancerName: aws.String("staging-alb-tf"),
				},
				&elbv2.LoadBalancer{
					LoadBalancerName: aws.String("production-alb-tf"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "omitted name",
			Src: `
resource "aws_security_group" "test" {
    name_prefix = "test-alb-tf"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "ALB is duplicate, but exists in state",
			Src: `
resource "aws_alb" "test" {
    name = "test-alb-tf"
}`,
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_alb.test": {
                    "type": "aws_alb",
                    "depends_on": [],
                    "primary": {
                        "id": "arn:aws:elasticloadbalancing:us-east-1:hogehoge:loadbalancer/app/test-alb-tf/fugafuga",
                        "attributes": {
                            "access_logs.#": "0",
                            "arn": "arn:aws:elasticloadbalancing:us-east-1:hogehoge:loadbalancer/app/test-alb-tf/fugafuga",
                            "arn_suffix": "app/test-alb-tf/fugafuga",
                            "dns_name": "test-alb-tf-hogehoge.us-east-1.elb.amazonaws.com",
                            "enable_deletion_protection": "false",
                            "id": "arn:aws:elasticloadbalancing:us-east-1:hogehoge:loadbalancer/app/test-alb-tf/fugafuga",
                            "idle_timeout": "60",
                            "internal": "false",
                            "name": "test-alb-tf",
                            "security_groups.#": "1",
                            "security_groups.123456789": "sg-1234abcd",
                            "subnets.#": "2",
                            "subnets.987654321": "subnet-1234abcd",
                            "subnets.234567892": "subnet-abcd1234",
                            "tags.%": "1",
                            "tags.Environment": "production",
                            "vpc_id": "vpc-1234abcd",
                            "zone_id": "ABCDEFGHIJ"
                        }
                    },
                    "provider": ""
                }
            }
        }
    ]
}
`,
			Response: []*elbv2.LoadBalancer{
				&elbv2.LoadBalancer{
					LoadBalancerName: aws.String("test-alb-tf"),
				},
				&elbv2.LoadBalancer{
					LoadBalancerName: aws.String("production-alb-tf"),
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
		elbv2mock := mock.NewMockELBV2API(ctrl)
		if tc.Response != nil {
			elbv2mock.EXPECT().DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{}).Return(&elbv2.DescribeLoadBalancersOutput{
				LoadBalancers: tc.Response,
			}, nil)
		}
		awsClient.Elbv2 = elbv2mock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsALBDuplicateNameDetector",
			tc.Src,
			tc.State,
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
