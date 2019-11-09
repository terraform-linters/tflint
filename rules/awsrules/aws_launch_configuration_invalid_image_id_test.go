package awsrules

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/client"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsLaunchConfigurationInvalidImageID_invalid(t *testing.T) {
	content := `
resource "aws_launch_configuration" "invalid" {
	image_id = "ami-1234abcd"
}`
	runner := tflint.TestRunner(t, map[string]string{"instances.tf": content})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ec2mock := client.NewMockEC2API(ctrl)
	ec2mock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: aws.StringSlice([]string{"ami-1234abcd"}),
	}).Return(&ec2.DescribeImagesOutput{
		Images: []*ec2.Image{},
	}, nil)
	runner.AwsClient.EC2 = ec2mock

	rule := NewAwsLaunchConfigurationInvalidImageIDRule()
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := tflint.Issues{
		{
			Rule:    NewAwsLaunchConfigurationInvalidImageIDRule(),
			Message: "\"ami-1234abcd\" is invalid image ID.",
			Range: hcl.Range{
				Filename: "instances.tf",
				Start:    hcl.Pos{Line: 3, Column: 13},
				End:      hcl.Pos{Line: 3, Column: 27},
			},
		},
	}

	tflint.AssertIssues(t, expected, runner.Issues)
}

func Test_AwsLaunchConfigurationInvalidImageID_valid(t *testing.T) {
	content := `
resource "aws_launch_configuration" "valid" {
	image_id = "ami-9ad76sd1"
}`
	runner := tflint.TestRunner(t, map[string]string{"instances.tf": content})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	rule := NewAwsLaunchConfigurationInvalidImageIDRule()
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := tflint.Issues{}
	tflint.AssertIssues(t, expected, runner.Issues)
}

func Test_AwsLaunchConfigurationInvalidImageID_error(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Request  *ec2.DescribeImagesInput
		Response error
		Error    tflint.Error
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
			Response: awserr.New(
				"MissingRegion",
				"could not find region configuration",
				nil,
			),
			Error: tflint.Error{
				Code:    tflint.ExternalAPIError,
				Level:   tflint.ErrorLevel,
				Message: "An error occurred while describing images; MissingRegion: could not find region configuration",
			},
		},
		{
			Name: "Unexpected error",
			Content: `
resource "aws_launch_configuration" "valid" {
	image_id = "ami-9ad76sd1"
}`,
			Request: &ec2.DescribeImagesInput{
				ImageIds: aws.StringSlice([]string{"ami-9ad76sd1"}),
			},
			Response: errors.New("Unexpected"),
			Error: tflint.Error{
				Code:    tflint.ExternalAPIError,
				Level:   tflint.ErrorLevel,
				Message: "An error occurred while describing images; Unexpected",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rule := NewAwsLaunchConfigurationInvalidImageIDRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"instances.tf": tc.Content})

		ec2mock := client.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeImages(tc.Request).Return(nil, tc.Response)
		runner.AwsClient.EC2 = ec2mock

		err := rule.Check(runner)
		tflint.AssertAppError(t, tc.Error, err)
	}
}

func Test_AwsLaunchConfigurationInvalidImageID_AMIError(t *testing.T) {
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
resource "aws_launch_configuration" "not_found" {
	image_id = "ami-9ad76sd1"
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
					Rule:    NewAwsLaunchConfigurationInvalidImageIDRule(),
					Message: "\"ami-9ad76sd1\" is invalid image ID.",
					Range: hcl.Range{
						Filename: "instances.tf",
						Start:    hcl.Pos{Line: 3, Column: 13},
						End:      hcl.Pos{Line: 3, Column: 27},
					},
				},
			},
			Error: false,
		},
		{
			Name: "malformed",
			Content: `
resource "aws_launch_configuration" "malformed" {
	image_id = "image-9ad76sd1"
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
					Rule:    NewAwsLaunchConfigurationInvalidImageIDRule(),
					Message: "\"image-9ad76sd1\" is invalid image ID.",
					Range: hcl.Range{
						Filename: "instances.tf",
						Start:    hcl.Pos{Line: 3, Column: 13},
						End:      hcl.Pos{Line: 3, Column: 29},
					},
				},
			},
			Error: false,
		},
		{
			Name: "unavailable",
			Content: `
resource "aws_launch_configuration" "unavailable" {
	image_id = "ami-1234567"
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
					Rule:    NewAwsLaunchConfigurationInvalidImageIDRule(),
					Message: "\"ami-1234567\" is invalid image ID.",
					Range: hcl.Range{
						Filename: "instances.tf",
						Start:    hcl.Pos{Line: 3, Column: 13},
						End:      hcl.Pos{Line: 3, Column: 26},
					},
				},
			},
			Error: false,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rule := NewAwsLaunchConfigurationInvalidImageIDRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"instances.tf": tc.Content})

		ec2mock := client.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeImages(tc.Request).Return(nil, tc.Response)
		runner.AwsClient.EC2 = ec2mock

		err := rule.Check(runner)
		if err != nil && !tc.Error {
			t.Fatalf("Failed `%s` test: unexpected error occurred: %s", tc.Name, err)
		}
		if err == nil && tc.Error {
			t.Fatalf("Failed `%s` test: expected to return an error, but nothing occurred", tc.Name)
		}

		tflint.AssertIssues(t, tc.Issues, runner.Issues)
	}
}
