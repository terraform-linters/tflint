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
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsELBInvalidSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected issue.Issues
	}{
		{
			Name: "security group is invalid",
			Content: `
resource "aws_elb" "balancer" {
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
			Expected: []*issue.Issue{
				{
					Detector: "aws_elb_invalid_security_group",
					Type:     "ERROR",
					Message:  "\"sg-1234abcd\" is invalid security group.",
					Line:     4,
					File:     "resource.tf",
				},
				{
					Detector: "aws_elb_invalid_security_group",
					Type:     "ERROR",
					Message:  "\"sg-abcd1234\" is invalid security group.",
					Line:     5,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "security group is valid",
			Content: `
resource "aws_elb" "balancer" {
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
			Expected: []*issue.Issue{},
		},
		{
			Name: "use list variable",
			Content: `
variable "security_groups" {
    default = ["sg-1234abcd", "sg-abcd1234"]
}

resource "aws_elb" "balancer" {
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
			Expected: []*issue.Issue{
				{
					Detector: "aws_elb_invalid_security_group",
					Type:     "ERROR",
					Message:  "\"sg-1234abcd\" is invalid security group.",
					Line:     7,
					File:     "resource.tf",
				},
				{
					Detector: "aws_elb_invalid_security_group",
					Type:     "ERROR",
					Message:  "\"sg-abcd1234\" is invalid security group.",
					Line:     7,
					File:     "resource.tf",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsELBInvalidSecurityGroup")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

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

		mod, diags := loader.Parser().LoadConfigDir(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}
		rule := NewAwsELBInvalidSecurityGroupRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
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
