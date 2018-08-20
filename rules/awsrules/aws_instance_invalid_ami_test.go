package awsrules

import (
	"errors"
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
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsInstanceInvalidAMI(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Image
		Expected issue.Issues
	}{
		{
			Name: "basic",
			Content: `
resource "aws_instance" "invalid" {
  ami = "ami-1234abcd"
}

resource "aws_instance" "valid" {
  ami = "ami-9ad76sd1"
}`,
			Response: []*ec2.Image{
				{
					ImageId: aws.String("ami-0c11b26d"),
				},
				{
					ImageId: aws.String("ami-9ad76sd1"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_ami",
					Type:     issue.ERROR,
					Message:  "\"ami-1234abcd\" is invalid AMI ID.",
					Line:     3,
					File:     "instances.tf",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidAMI")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/instances.tf", []byte(tc.Content), os.ModePerm)
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

		runner := tflint.NewRunner(config.Init(), cfg, map[string]*terraform.InputValue{})
		rule := NewAwsInstanceInvalidAMIRule()

		ec2mock := mock.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{}).Return(&ec2.DescribeImagesOutput{
			Images: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = ec2mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}

func Test_AwsInstanceInvalidAMI_error(t *testing.T) {
	cases := []struct {
		Name       string
		Content    string
		Response   error
		ErrorCode  int
		ErrorLevel int
		ErrorText  string
	}{
		{
			Name: "AWS API error",
			Content: `
resource "aws_instance" "valid" {
  ami = "ami-9ad76sd1"
}`,
			Response:   errors.New("MissingRegion: could not find region configuration"),
			ErrorCode:  tflint.ExternalAPIError,
			ErrorLevel: tflint.ErrorLevel,
			ErrorText:  "An error occurred while describing images; MissingRegion: could not find region configuration",
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidAMI_error")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/instances.tf", []byte(tc.Content), os.ModePerm)
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

		runner := tflint.NewRunner(config.Init(), cfg, map[string]*terraform.InputValue{})
		rule := NewAwsInstanceInvalidAMIRule()

		ec2mock := mock.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{}).Return(nil, tc.Response)
		runner.AwsClient.EC2 = ec2mock

		err = rule.Check(runner)
		if appErr, ok := err.(*tflint.Error); ok {
			if appErr == nil {
				t.Fatalf("Failed `%s` test: expected err is `%s`, but nothing occurred", tc.Name, tc.ErrorText)
			}
			if appErr.Code != tc.ErrorCode {
				t.Fatalf("Failed `%s` test: expected error code is `%d`, but get `%d`", tc.Name, tc.ErrorCode, appErr.Code)
			}
			if appErr.Level != tc.ErrorLevel {
				t.Fatalf("Failed `%s` test: expected error level is `%d`, but get `%d`", tc.Name, tc.ErrorLevel, appErr.Level)
			}
			if appErr.Error() != tc.ErrorText {
				t.Fatalf("Failed `%s` test: expected error is `%s`, but get `%s`", tc.Name, tc.ErrorText, appErr.Error())
			}
		} else {
			t.Fatalf("Failed `%s` test: unexpected error occurred: %s", tc.Name, err)
		}
	}
}
