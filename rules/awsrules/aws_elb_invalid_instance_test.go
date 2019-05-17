package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsELBInvalidInstance(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Instance
		Expected issue.Issues
	}{
		{
			Name: "Instance is invalid",
			Content: `
resource "aws_elb" "balancer" {
    instances = [
        "i-1234abcd",
        "i-abcd1234",
    ]
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-12345678"),
				},
				{
					InstanceId: aws.String("i-abcdefgh"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_elb_invalid_instance",
					Type:     "ERROR",
					Message:  "\"i-1234abcd\" is invalid instance.",
					Line:     4,
					File:     "resource.tf",
				},
				{
					Detector: "aws_elb_invalid_instance",
					Type:     "ERROR",
					Message:  "\"i-abcd1234\" is invalid instance.",
					Line:     5,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "Instance is valid",
			Content: `
resource "aws_elb" "balancer" {
    instances = [
        "i-1234abcd",
        "i-abcd1234",
    ]
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-1234abcd"),
				},
				{
					InstanceId: aws.String("i-abcd1234"),
				},
			},
			Expected: []*issue.Issue{},
		},
		{
			Name: "use list variable",
			Content: `
variable "instances" {
    default = ["i-1234abcd", "i-abcd1234"]
}

resource "aws_elb" "balancer" {
    instances = "${var.instances}"
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-12345678"),
				},
				{
					InstanceId: aws.String("i-abcdefgh"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_elb_invalid_instance",
					Type:     "ERROR",
					Message:  "\"i-1234abcd\" is invalid instance.",
					Line:     7,
					File:     "resource.tf",
				},
				{
					Detector: "aws_elb_invalid_instance",
					Type:     "ERROR",
					Message:  "\"i-abcd1234\" is invalid instance.",
					Line:     7,
					File:     "resource.tf",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsELBInvalidInstance")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(dir)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		rule := NewAwsELBInvalidInstanceRule()

		mock := mock.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeInstances(&ec2.DescribeInstancesInput{}).Return(&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: tc.Response,
				},
			},
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
