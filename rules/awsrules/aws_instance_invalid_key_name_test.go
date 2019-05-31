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

func Test_AwsInstanceInvalidKeyName(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.KeyPairInfo
		Expected issue.Issues
	}{
		{
			Name: "Key name is invalid",
			Content: `
resource "aws_instance" "web" {
    key_name = "foo"
}`,
			Response: []*ec2.KeyPairInfo{
				{
					KeyName: aws.String("hogehoge"),
				},
				{
					KeyName: aws.String("fugafuga"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_key_name",
					Type:     "ERROR",
					Message:  "\"foo\" is invalid key name.",
					Line:     3,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "Key name is valid",
			Content: `
resource "aws_instance" "web" {
    key_name = "foo"
}`,
			Response: []*ec2.KeyPairInfo{
				{
					KeyName: aws.String("foo"),
				},
				{
					KeyName: aws.String("bar"),
				},
			},
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidKeyName")
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
		rule := NewAwsInstanceInvalidKeyNameRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeKeyPairs(&ec2.DescribeKeyPairsInput{}).Return(&ec2.DescribeKeyPairsOutput{
			KeyPairs: tc.Response,
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
