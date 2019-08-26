package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
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

func Test_AwsDBInstanceInvalidParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*rds.DBParameterGroup
		Expected tflint.Issues
	}{
		{
			Name: "parameter_group_name is invalid",
			Content: `
resource "aws_db_instance" "mysql" {
    parameter_group_name = "app-server"
}`,
			Response: []*rds.DBParameterGroup{
				{
					DBParameterGroupName: aws.String("app-server1"),
				},
				{
					DBParameterGroupName: aws.String("app-server2"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidParameterGroupRule(),
					Message: "\"app-server\" is invalid parameter group name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 40},
					},
				},
			},
		},
		{
			Name: "parameter_group_name is valid",
			Content: `
resource "aws_db_instance" "mysql" {
    parameter_group_name = "app-server"
}`,
			Response: []*rds.DBParameterGroup{
				{
					DBParameterGroupName: aws.String("app-server1"),
				},
				{
					DBParameterGroupName: aws.String("app-server2"),
				},
				{
					DBParameterGroupName: aws.String("app-server"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceInvalidParameterGroup")
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
		rule := NewAwsDBInstanceInvalidParameterGroupRule()

		mock := client.NewMockRDSAPI(ctrl)
		mock.EXPECT().DescribeDBParameterGroups(&rds.DescribeDBParameterGroupsInput{}).Return(&rds.DescribeDBParameterGroupsOutput{
			DBParameterGroups: tc.Response,
		}, nil)
		runner.AwsClient.RDS = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsDBInstanceInvalidParameterGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}
