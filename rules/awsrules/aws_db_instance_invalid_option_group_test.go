package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsDBInstanceInvalidOptionGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*rds.OptionGroup
		Expected issue.Issues
	}{
		{
			Name: "option_group is invalid",
			Content: `
resource "aws_db_instance" "mysql" {
    option_group_name = "app-server"
}`,
			Response: []*rds.OptionGroup{
				{
					OptionGroupName: aws.String("app-server1"),
				},
				{
					OptionGroupName: aws.String("app-server2"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_db_instance_invalid_option_group",
					Type:     "ERROR",
					Message:  "\"app-server\" is invalid option group name.",
					Line:     3,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "option_group is valid",
			Content: `
resource "aws_db_instance" "mysql" {
    option_group_name = "app-server"
}`,
			Response: []*rds.OptionGroup{
				{
					OptionGroupName: aws.String("app-server1"),
				},
				{
					OptionGroupName: aws.String("app-server2"),
				},
				{
					OptionGroupName: aws.String("app-server"),
				},
			},
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceInvalidOptionGroup")
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
		rule := NewAwsDBInstanceInvalidOptionGroupRule()

		mock := client.NewMockRDSAPI(ctrl)
		mock.EXPECT().DescribeOptionGroups(&rds.DescribeOptionGroupsInput{}).Return(&rds.DescribeOptionGroupsOutput{
			OptionGroupsList: tc.Response,
		}, nil)
		runner.AwsClient.RDS = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
