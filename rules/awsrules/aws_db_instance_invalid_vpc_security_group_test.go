package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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

func Test_AwsDBInstanceInvalidVPCSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected tflint.Issues
	}{
		{
			Name: "security group is invalid",
			Content: `
resource "aws_db_instance" "mysql" {
    vpc_security_group_ids = [
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidVPCSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsDBInstanceInvalidVPCSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 22},
					},
				},
			},
		},
		{
			Name: "security group is valid",
			Content: `
resource "aws_db_instance" "mysql" {
    vpc_security_group_ids = [
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
			Expected: tflint.Issues{},
		},
		{
			Name: "use list variable",
			Content: `
variable "security_groups" {
   default = ["sg-1234abcd", "sg-abcd1234"]
}

resource "aws_db_instance" "mysql" {
    vpc_security_group_ids = "${var.security_groups}"
}`,
			Response: []*ec2.SecurityGroup{
				{
					GroupId: aws.String("sg-12345678"),
				},
				{
					GroupId: aws.String("sg-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidVPCSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 30},
						End:      hcl.Pos{Line: 7, Column: 54},
					},
				},
				{
					Rule:    NewAwsDBInstanceInvalidVPCSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 30},
						End:      hcl.Pos{Line: 7, Column: 54},
					},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceInvalidVpcSecurityGroup")
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
		rule := NewAwsDBInstanceInvalidVPCSecurityGroupRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsDBInstanceInvalidVPCSecurityGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}
