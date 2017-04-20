package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsELBDuplicateName(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		State    string
		Response []*elb.LoadBalancerDescription
		Issues   []*issue.Issue
	}{
		{
			Name: "ELB name is duplicate",
			Src: `
resource "aws_elb" "test" {
    name = "test-elb-tf"
}`,
			Response: []*elb.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("test-elb-tf"),
				},
				{
					LoadBalancerName: aws.String("production-elb-tf"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"test-elb-tf\" is duplicate name. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "ELB name is unique",
			Src: `
resource "aws_elb" "test" {
    name = "test-elb-tf"
}`,
			Response: []*elb.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("staging-elb-tf"),
				},
				{
					LoadBalancerName: aws.String("production-elb-tf"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "omitted name",
			Src: `
resource "aws_elb" "test" {
    instances = ["i-12345abcdf"]
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "ELB is duplicate, but exists in state",
			Src: `
resource "aws_elb" "test" {
    name = "test-elb-tf"
}`,
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_elb.test": {
                    "type": "aws_elb",
                    "depends_on": [],
                    "primary": {
                        "id": "test-elb-tf",
                        "attributes": {
                            "access_logs.#": "0",
                            "availability_zones.#": "2",
                            "availability_zones.123456678": "us-east-1a",
                            "availability_zones.527875093": "us-east-1b",
                            "connection_draining": "false",
                            "connection_draining_timeout": "300",
                            "cross_zone_load_balancing": "true",
                            "dns_name": "test-elb-tf-12345678.us-east-1.elb.amazonaws.com",
                            "health_check.#": "1",
                            "health_check.0.healthy_threshold": "10",
                            "health_check.0.interval": "30",
                            "health_check.0.target": "TCP:8000",
                            "health_check.0.timeout": "5",
                            "health_check.0.unhealthy_threshold": "2",
                            "id": "test-elb-tf",
                            "idle_timeout": "60",
                            "instances.#": "0",
                            "internal": "false",
                            "listener.#": "1",
                            "listener.206423021.instance_port": "8000",
                            "listener.206423021.instance_protocol": "http",
                            "listener.206423021.lb_port": "80",
                            "listener.206423021.lb_protocol": "http",
                            "listener.206423021.ssl_certificate_id": "",
                            "name": "test-elb-tf",
                            "security_groups.#": "1",
                            "security_groups.3963419045": "sg-1234abcd",
                            "source_security_group": "988578293/default",
                            "source_security_group_id": "sg-abcd1234",
                            "subnets.#": "2",
                            "subnets.2342537193": "subnet-1234abcd",
                            "subnets.3798310056": "subnet-abcd1234",
                            "tags.%": "0",
                            "zone_id": "KAHN32700A"
                        }
                    },
                    "provider": ""
                }
            }
        }
    ]
}
`,
			Response: []*elb.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("test-elb-tf"),
				},
				{
					LoadBalancerName: aws.String("production-elb-tf"),
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
		elbmock := mock.NewMockELBAPI(ctrl)
		elbmock.EXPECT().DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{}).Return(&elb.DescribeLoadBalancersOutput{
			LoadBalancerDescriptions: tc.Response,
		}, nil)
		awsClient.Elb = elbmock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsELBDuplicateNameDetector",
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
