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
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsLaunchConfigurationInvalidImageID_invalid(t *testing.T) {
	dir, err := ioutil.TempDir("", "AwsLaunchConfigurationInvalidImageID_invalid")
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

	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	content := `
resource "aws_launch_configuration" "invalid" {
	image_id = "ami-1234abcd"
}`
	err = ioutil.WriteFile(dir+"/instances.tf", []byte(content), os.ModePerm)
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

	runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
	rule := NewAwsLaunchConfigurationInvalidImageIDRule()

	ec2mock := client.NewMockEC2API(ctrl)
	ec2mock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: aws.StringSlice([]string{"ami-1234abcd"}),
	}).Return(&ec2.DescribeImagesOutput{
		Images: []*ec2.Image{},
	}, nil)
	runner.AwsClient.EC2 = ec2mock

	if err = rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := issue.Issues{
		{
			Detector: "aws_launch_configuration_invalid_image_id",
			Type:     issue.ERROR,
			Message:  "\"ami-1234abcd\" is invalid image ID.",
			Line:     3,
			File:     "instances.tf",
		},
	}
	if !cmp.Equal(expected, runner.Issues) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, runner.Issues))
	}
}

func Test_AwsLaunchConfigurationInvalidImageID_valid(t *testing.T) {
	dir, err := ioutil.TempDir("", "AwsLaunchConfigurationInvalidImageID_invalid")
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

	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	content := `
resource "aws_launch_configuration" "valid" {
	image_id = "ami-9ad76sd1"
}`
	err = ioutil.WriteFile(dir+"/instances.tf", []byte(content), os.ModePerm)
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

	runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
	rule := NewAwsLaunchConfigurationInvalidImageIDRule()

	ec2mock := client.NewMockEC2API(ctrl)
	ec2mock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: aws.StringSlice([]string{"ami-9ad76sd1"}),
	}).Return(&ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				ImageId: aws.String("ami-9ad76sd1"),
			},
		},
	}, nil)
	runner.AwsClient.EC2 = ec2mock

	if err = rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := issue.Issues{}
	if !cmp.Equal(expected, runner.Issues) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, runner.Issues))
	}
}

func Test_AwsLaunchConfigurationInvalidImageID_error(t *testing.T) {
	cases := []struct {
		Name       string
		Content    string
		Request    *ec2.DescribeImagesInput
		Response   error
		ErrorCode  int
		ErrorLevel int
		ErrorText  string
	}{
		{
			Name: "AWS API error",
			Content: `
resource "aws_launch_configuration" "valid" {
  image_id = "ami-9ad76sd1"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"ami-9ad76sd1"}),
			},
			Response:   errors.New("MissingRegion: could not find region configuration"),
			ErrorCode:  tflint.ExternalAPIError,
			ErrorLevel: tflint.ErrorLevel,
			ErrorText:  "An error occurred while describing images; MissingRegion: could not find region configuration",
		},
	}

	dir, err := ioutil.TempDir("", "AwsLaunchConfigurationInvalidImageID_error")
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

		mod, diags := loader.Parser().LoadConfigDir(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		rule := NewAwsLaunchConfigurationInvalidImageIDRule()

		ec2mock := client.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeImages(tc.Request).Return(nil, tc.Response)
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
