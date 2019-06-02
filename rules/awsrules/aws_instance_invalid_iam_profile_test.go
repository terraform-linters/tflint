package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsInstanceInvalidIAMProfile(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*iam.InstanceProfile
		Expected issue.Issues
	}{
		{
			Name: "iam_instance_profile is invalid",
			Content: `
resource "aws_instance" "web" {
    iam_instance_profile = "app-server"
}`,
			Response: []*iam.InstanceProfile{
				{
					InstanceProfileName: aws.String("app-server1"),
				},
				{
					InstanceProfileName: aws.String("app-server2"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_iam_profile",
					Type:     "ERROR",
					Message:  "\"app-server\" is invalid IAM profile name.",
					Line:     3,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "iam_instance_profile is valid",
			Content: `
resource "aws_instance" "web" {
    iam_instance_profile = "app-server"
}`,
			Response: []*iam.InstanceProfile{
				{
					InstanceProfileName: aws.String("app-server1"),
				},
				{
					InstanceProfileName: aws.String("app-server2"),
				},
				{
					InstanceProfileName: aws.String("app-server"),
				},
			},
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidIamProfile")
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

		runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		rule := NewAwsInstanceInvalidIAMProfileRule()

		mock := client.NewMockIAMAPI(ctrl)
		mock.EXPECT().ListInstanceProfiles(&iam.ListInstanceProfilesInput{}).Return(&iam.ListInstanceProfilesOutput{
			InstanceProfiles: tc.Response,
		}, nil)
		runner.AwsClient.IAM = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
