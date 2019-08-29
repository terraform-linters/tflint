package awsrules

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsInstanceInvalidAMI_invalid(t *testing.T) {
	dir, err := ioutil.TempDir("", "AwsInstanceInvalidAMI_invalid")
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
resource "aws_instance" "invalid" {
	ami = "ami-1234abcd"
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

	runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}
	rule := NewAwsInstanceInvalidAMIRule()

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

	expected := tflint.Issues{
		{
			Rule:    NewAwsInstanceInvalidAMIRule(),
			Message: "\"ami-1234abcd\" is invalid AMI ID.",
			Range: hcl.Range{
				Filename: "instances.tf",
				Start:    hcl.Pos{Line: 3, Column: 8},
				End:      hcl.Pos{Line: 3, Column: 22},
			},
		},
	}
	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(AwsInstanceInvalidAMIRule{}),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, runner.Issues, opts...) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, runner.Issues, opts...))
	}
}

func Test_AwsInstanceInvalidAMI_valid(t *testing.T) {
	dir, err := ioutil.TempDir("", "AwsInstanceInvalidAMI_invalid")
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
resource "aws_instance" "valid" {
	ami = "ami-9ad76sd1"
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

	runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}
	rule := NewAwsInstanceInvalidAMIRule()

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

	expected := tflint.Issues{}
	if !cmp.Equal(expected, runner.Issues) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, runner.Issues))
	}
}

func Test_AwsInstanceInvalidAMI_error(t *testing.T) {
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
resource "aws_instance" "valid" {
  ami = "ami-9ad76sd1"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"ami-9ad76sd1"}),
			},
			Response: awserr.New(
				"MissingRegion",
				"could not find region configuration",
				nil,
			),
			ErrorCode:  tflint.ExternalAPIError,
			ErrorLevel: tflint.ErrorLevel,
			ErrorText:  "An error occurred while describing images; MissingRegion: could not find region configuration",
		},
		{
			Name: "Unexpected error",
			Content: `
resource "aws_instance" "valid" {
  ami = "ami-9ad76sd1"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"ami-9ad76sd1"}),
			},
			Response:   errors.New("Unexpected"),
			ErrorCode:  tflint.ExternalAPIError,
			ErrorLevel: tflint.ErrorLevel,
			ErrorText:  "An error occurred while describing images; Unexpected",
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidAMI_error")
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
		err := ioutil.WriteFile(dir+"/instances.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		loader, err := configload.NewLoader(&configload.Config{})
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
		rule := NewAwsInstanceInvalidAMIRule()

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

func Test_AwsInstanceInvalidAMI_AMIError(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Request  *ec2.DescribeImagesInput
		Response error
		Issues   tflint.Issues
		Error    bool
	}{
		{
			Name: "not found",
			Content: `
resource "aws_instance" "not_found" {
	ami = "ami-9ad76sd1"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"ami-9ad76sd1"}),
			},
			Response: awserr.New(
				"InvalidAMIID.NotFound",
				"The image id '[ami-9ad76sd1]' does not exist",
				nil,
			),
			Issues: tflint.Issues{
				{
					Rule:    NewAwsInstanceInvalidAMIRule(),
					Message: "\"ami-9ad76sd1\" is invalid AMI ID.",
					Range: hcl.Range{
						Filename: "instances.tf",
						Start:    hcl.Pos{Line: 3, Column: 8},
						End:      hcl.Pos{Line: 3, Column: 22},
					},
				},
			},
			Error: false,
		},
		{
			Name: "malformed",
			Content: `
resource "aws_instance" "malformed" {
	ami = "image-9ad76sd1"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"image-9ad76sd1"}),
			},
			Response: awserr.New(
				"InvalidAMIID.Malformed",
				"Invalid id: \"image-9ad76sd1\" (expecting \"ami-...\")",
				nil,
			),
			Issues: tflint.Issues{
				{
					Rule:    NewAwsInstanceInvalidAMIRule(),
					Message: "\"image-9ad76sd1\" is invalid AMI ID.",
					Range: hcl.Range{
						Filename: "instances.tf",
						Start:    hcl.Pos{Line: 3, Column: 8},
						End:      hcl.Pos{Line: 3, Column: 24},
					},
				},
			},
			Error: false,
		},
		{
			Name: "unavailable",
			Content: `
resource "aws_instance" "unavailable" {
	ami = "ami-1234567"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"ami-1234567"}),
			},
			Response: awserr.New(
				"InvalidAMIID.Unavailable",
				"The image ID: 'ami-1234567' is no longer available",
				nil,
			),
			Issues: tflint.Issues{
				{
					Rule:    NewAwsInstanceInvalidAMIRule(),
					Message: "\"ami-1234567\" is invalid AMI ID.",
					Range: hcl.Range{
						Filename: "instances.tf",
						Start:    hcl.Pos{Line: 3, Column: 8},
						End:      hcl.Pos{Line: 3, Column: 21},
					},
				},
			},
			Error: false,
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidAMI_AMIError")
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
		err := ioutil.WriteFile(dir+"/instances.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		loader, err := configload.NewLoader(&configload.Config{})
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
		rule := NewAwsInstanceInvalidAMIRule()

		ec2mock := client.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeImages(tc.Request).Return(nil, tc.Response)
		runner.AwsClient.EC2 = ec2mock

		err = rule.Check(runner)
		if err != nil && !tc.Error {
			t.Fatalf("Failed `%s` test: unexpected error occurred: %s", tc.Name, err)
		}
		if err == nil && tc.Error {
			t.Fatalf("Failed `%s` test: expected to return an error, but nothing occurred", tc.Name)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsInstanceInvalidAMIRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Issues, runner.Issues, opts...) {
			t.Fatalf("Failed `%s` test: expected issues are not matched:\n %s\n", tc.Name, cmp.Diff(tc.Issues, runner.Issues, opts...))
		}
	}
}
