// Migration tests are tests to verify that compatibility is kept when migrating existing rules.
// If you need to change these tests, you can delete it instead.

package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsALBInvalidSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected tflint.Issues
	}{
		{
			Name: "security group is invalid",
			Content: `
resource "aws_alb" "balancer" {
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
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
resource "aws_alb" "balancer" {
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
			Expected: tflint.Issues{},
		},
		{
			Name: "use list variables",
			Content: `
variable "security_groups" {
    default = ["sg-1234abcd", "sg-abcd1234"]
}

resource "aws_alb" "balancer" {
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 23},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
				},
				{
					Rule:    NewAwsALBInvalidSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 23},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsAlbInvalidSecurityGroup")
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
		rule := NewAwsALBInvalidSecurityGroupRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsALBInvalidSecurityGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsALBInvalidSubnet(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Subnet
		Expected tflint.Issues
	}{
		{
			Name: "subnet ID is invalid",
			Content: `
resource "aws_alb" "balancer" {
    subnets = [
        "subnet-1234abcd",
        "subnet-abcd1234",
    ]
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-12345678"),
				},
				{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsALBInvalidSubnetRule(),
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 26},
					},
				},
				{
					Rule:    NewAwsALBInvalidSubnetRule(),
					Message: "\"subnet-abcd1234\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 26},
					},
				},
			},
		},
		{
			Name: "subnet ID is valid",
			Content: `
resource "aws_alb" "balancer" {
    subnets = [
        "subnet-1234abcd",
        "subnet-abcd1234",
    ]
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-1234abcd"),
				},
				{
					SubnetId: aws.String("subnet-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
		{
			Name: "use list variables",
			Content: `
variable "subnets" {
    default = ["subnet-1234abcd", "subnet-abcd1234"]
}

resource "aws_alb" "balancer" {
    subnets = "${var.subnets}"
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-1234abcd"),
				},
				{
					SubnetId: aws.String("subnet-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsALBInvalidSubnet")
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
		rule := NewAwsALBInvalidSubnetRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{}).Return(&ec2.DescribeSubnetsOutput{
			Subnets: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsALBInvalidSubnetRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsDBInstanceInvalidDBSubnetGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*rds.DBSubnetGroup
		Expected tflint.Issues
	}{
		{
			Name: "db_subnet_group_name is invalid",
			Content: `
resource "aws_db_instance" "mysql" {
    db_subnet_group_name = "app-server"
}`,
			Response: []*rds.DBSubnetGroup{
				{
					DBSubnetGroupName: aws.String("app-server1"),
				},
				{
					DBSubnetGroupName: aws.String("app-server2"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidDBSubnetGroupRule(),
					Message: "\"app-server\" is invalid DB subnet group name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 40},
					},
				},
			},
		},
		{
			Name: "db_subnet_group_name is valid",
			Content: `
resource "aws_db_instance" "mysql" {
    db_subnet_group_name = "app-server"
}`,
			Response: []*rds.DBSubnetGroup{
				{
					DBSubnetGroupName: aws.String("app-server1"),
				},
				{
					DBSubnetGroupName: aws.String("app-server2"),
				},
				{
					DBSubnetGroupName: aws.String("app-server"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceInvalidDBSubnetGroup")
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
		rule := NewAwsDBInstanceInvalidDBSubnetGroupRule()

		mock := client.NewMockRDSAPI(ctrl)
		mock.EXPECT().DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{}).Return(&rds.DescribeDBSubnetGroupsOutput{
			DBSubnetGroups: tc.Response,
		}, nil)
		runner.AwsClient.RDS = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsDBInstanceInvalidDBSubnetGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsDBInstanceInvalidOptionGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*rds.OptionGroup
		Expected tflint.Issues
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidOptionGroupRule(),
					Message: "\"app-server\" is invalid option group name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 25},
						End:      hcl.Pos{Line: 3, Column: 37},
					},
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
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceInvalidOptionGroup")
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
		rule := NewAwsDBInstanceInvalidOptionGroupRule()

		mock := client.NewMockRDSAPI(ctrl)
		mock.EXPECT().DescribeOptionGroups(&rds.DescribeOptionGroupsInput{}).Return(&rds.DescribeOptionGroupsOutput{
			OptionGroupsList: tc.Response,
		}, nil)
		runner.AwsClient.RDS = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsDBInstanceInvalidOptionGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

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
					Rule:    NewAwsDBInstanceInvalidVpcSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsDBInstanceInvalidVpcSecurityGroupRule(),
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
					Rule:    NewAwsDBInstanceInvalidVpcSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 30},
						End:      hcl.Pos{Line: 7, Column: 54},
					},
				},
				{
					Rule:    NewAwsDBInstanceInvalidVpcSecurityGroupRule(),
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
		rule := NewAwsDBInstanceInvalidVpcSecurityGroupRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsDBInstanceInvalidVpcSecurityGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsElastiCacheClusterInvalidParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*elasticache.CacheParameterGroup
		Expected tflint.Issues
	}{
		{
			Name: "parameter_group_name is invalid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    parameter_group_name = "app-server"
}`,
			Response: []*elasticache.CacheParameterGroup{
				{
					CacheParameterGroupName: aws.String("app-server1"),
				},
				{
					CacheParameterGroupName: aws.String("app-server2"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsElastiCacheClusterInvalidParameterGroupRule(),
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
resource "aws_elasticache_cluster" "redis" {
    parameter_group_name = "app-server"
}`,
			Response: []*elasticache.CacheParameterGroup{
				{
					CacheParameterGroupName: aws.String("app-server1"),
				},
				{
					CacheParameterGroupName: aws.String("app-server2"),
				},
				{
					CacheParameterGroupName: aws.String("app-server"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsElasticacheClusterInvalidParameterGroup")
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
		rule := NewAwsElastiCacheClusterInvalidParameterGroupRule()

		mock := client.NewMockElastiCacheAPI(ctrl)
		mock.EXPECT().DescribeCacheParameterGroups(&elasticache.DescribeCacheParameterGroupsInput{}).Return(&elasticache.DescribeCacheParameterGroupsOutput{
			CacheParameterGroups: tc.Response,
		}, nil)
		runner.AwsClient.ElastiCache = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsElastiCacheClusterInvalidParameterGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsElastiCacheClusterInvalidSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected tflint.Issues
	}{
		{
			Name: "security group is invalid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    security_group_ids = [
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
					Rule:    NewAwsElastiCacheClusterInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsElastiCacheClusterInvalidSecurityGroupRule(),
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
resource "aws_elasticache_cluster" "redis" {
    security_group_ids = [
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

resource "aws_elasticache_cluster" "redis" {
    security_group_ids = "${var.security_groups}"
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
					Rule:    NewAwsElastiCacheClusterInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 26},
						End:      hcl.Pos{Line: 7, Column: 50},
					},
				},
				{
					Rule:    NewAwsElastiCacheClusterInvalidSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 26},
						End:      hcl.Pos{Line: 7, Column: 50},
					},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsElastiCacheClusterInvalidSecurityGroup")
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
		rule := NewAwsElastiCacheClusterInvalidSecurityGroupRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsElastiCacheClusterInvalidSecurityGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsElastiCacheClusterInvalidSubnetGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*elasticache.CacheSubnetGroup
		Expected tflint.Issues
	}{
		{
			Name: "parameter_group_name is invalid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    subnet_group_name = "app-server"
}`,
			Response: []*elasticache.CacheSubnetGroup{
				{
					CacheSubnetGroupName: aws.String("app-server1"),
				},
				{
					CacheSubnetGroupName: aws.String("app-server2"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsElastiCacheClusterInvalidSubnetGroupRule(),
					Message: "\"app-server\" is invalid subnet group name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 25},
						End:      hcl.Pos{Line: 3, Column: 37},
					},
				},
			},
		},
		{
			Name: "parameter_group_name is valid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    subnet_group_name = "app-server"
}`,
			Response: []*elasticache.CacheSubnetGroup{
				{
					CacheSubnetGroupName: aws.String("app-server1"),
				},
				{
					CacheSubnetGroupName: aws.String("app-server2"),
				},
				{
					CacheSubnetGroupName: aws.String("app-server"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsElastiCacheClusterInvalidSubnetGroup")
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
		rule := NewAwsElastiCacheClusterInvalidSubnetGroupRule()

		mock := client.NewMockElastiCacheAPI(ctrl)
		mock.EXPECT().DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{}).Return(&elasticache.DescribeCacheSubnetGroupsOutput{
			CacheSubnetGroups: tc.Response,
		}, nil)
		runner.AwsClient.ElastiCache = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsElastiCacheClusterInvalidSubnetGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsELBInvalidInstance(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Instance
		Expected tflint.Issues
	}{
		{
			Name: "Instance is invalid",
			Content: `
resource "aws_elb" "balancer" {
    instances = [
        "i-1234abcd",
        "i-abcd1234",
    ]
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-12345678"),
				},
				{
					InstanceId: aws.String("i-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsELBInvalidInstanceRule(),
					Message: "\"i-1234abcd\" is invalid instance.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 21},
					},
				},
				{
					Rule:    NewAwsELBInvalidInstanceRule(),
					Message: "\"i-abcd1234\" is invalid instance.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 21},
					},
				},
			},
		},
		{
			Name: "Instance is valid",
			Content: `
resource "aws_elb" "balancer" {
    instances = [
        "i-1234abcd",
        "i-abcd1234",
    ]
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-1234abcd"),
				},
				{
					InstanceId: aws.String("i-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
		{
			Name: "use list variable",
			Content: `
variable "instances" {
    default = ["i-1234abcd", "i-abcd1234"]
}

resource "aws_elb" "balancer" {
    instances = "${var.instances}"
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-12345678"),
				},
				{
					InstanceId: aws.String("i-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsELBInvalidInstanceRule(),
					Message: "\"i-1234abcd\" is invalid instance.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 17},
						End:      hcl.Pos{Line: 7, Column: 35},
					},
				},
				{
					Rule:    NewAwsELBInvalidInstanceRule(),
					Message: "\"i-abcd1234\" is invalid instance.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 17},
						End:      hcl.Pos{Line: 7, Column: 35},
					},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsELBInvalidInstance")
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
		rule := NewAwsELBInvalidInstanceRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeInstances(&ec2.DescribeInstancesInput{}).Return(&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: tc.Response,
				},
			},
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsELBInvalidInstanceRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsELBInvalidSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected tflint.Issues
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsELBInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsELBInvalidSecurityGroupRule(),
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsELBInvalidSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 23},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
				},
				{
					Rule:    NewAwsELBInvalidSecurityGroupRule(),
					Message: "\"sg-abcd1234\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 23},
						End:      hcl.Pos{Line: 7, Column: 47},
					},
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

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsELBInvalidSecurityGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsELBInvalidSubnet(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Subnet
		Expected tflint.Issues
	}{
		{
			Name: "Subnet ID is invalid",
			Content: `
resource "aws_elb" "balancer" {
    subnets = [
        "subnet-1234abcd",
        "subnet-abcd1234",
    ]
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-12345678"),
				},
				{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsELBInvalidSubnetRule(),
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 26},
					},
				},
				{
					Rule:    NewAwsELBInvalidSubnetRule(),
					Message: "\"subnet-abcd1234\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 5, Column: 9},
						End:      hcl.Pos{Line: 5, Column: 26},
					},
				},
			},
		},
		{
			Name: "Subnet ID is valid",
			Content: `
resource "aws_elb" "balancer" {
    subnets = [
        "subnet-1234abcd",
        "subnet-abcd1234",
    ]
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-1234abcd"),
				},
				{
					SubnetId: aws.String("subnet-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
		{
			Name: "use list variable",
			Content: `
variable "subnets" {
    default = ["subnet-1234abcd", "subnet-abcd1234"]
}

resource "aws_elb" "balancer" {
    subnets = "${var.subnets}"
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-12345678"),
				},
				{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsELBInvalidSubnetRule(),
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 15},
						End:      hcl.Pos{Line: 7, Column: 31},
					},
				},
				{
					Rule:    NewAwsELBInvalidSubnetRule(),
					Message: "\"subnet-abcd1234\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 15},
						End:      hcl.Pos{Line: 7, Column: 31},
					},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsELBInvalidSubnet")
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
		rule := NewAwsELBInvalidSubnetRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{}).Return(&ec2.DescribeSubnetsOutput{
			Subnets: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsELBInvalidSubnetRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsInstanceInvalidIAMProfile(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*iam.InstanceProfile
		Expected tflint.Issues
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsInstanceInvalidIAMProfileRule(),
					Message: "\"app-server\" is invalid IAM profile name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 40},
					},
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
			Expected: tflint.Issues{},
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

		runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}
		rule := NewAwsInstanceInvalidIAMProfileRule()

		mock := client.NewMockIAMAPI(ctrl)
		mock.EXPECT().ListInstanceProfiles(&iam.ListInstanceProfilesInput{}).Return(&iam.ListInstanceProfilesOutput{
			InstanceProfiles: tc.Response,
		}, nil)
		runner.AwsClient.IAM = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsInstanceInvalidIAMProfileRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsInstanceInvalidKeyName(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.KeyPairInfo
		Expected tflint.Issues
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsInstanceInvalidKeyNameRule(),
					Message: "\"foo\" is invalid key name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 16},
						End:      hcl.Pos{Line: 3, Column: 21},
					},
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
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidKeyName")
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
		rule := NewAwsInstanceInvalidKeyNameRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeKeyPairs(&ec2.DescribeKeyPairsInput{}).Return(&ec2.DescribeKeyPairsOutput{
			KeyPairs: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsInstanceInvalidKeyNameRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsInstanceInvalidSubnet(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Subnet
		Expected tflint.Issues
	}{
		{
			Name: "Subnet ID is invalid",
			Content: `
resource "aws_instance" "web" {
    subnet_id = "subnet-1234abcd"
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-12345678"),
				},
				{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsInstanceInvalidSubnetRule(),
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 17},
						End:      hcl.Pos{Line: 3, Column: 34},
					},
				},
			},
		},
		{
			Name: "Subnet ID is valid",
			Content: `
resource "aws_instance" "web" {
    subnet_id = "subnet-1234abcd"
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-1234abcd"),
				},
				{
					SubnetId: aws.String("subnet-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidSubnet")
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
		rule := NewAwsInstanceInvalidSubnetRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{}).Return(&ec2.DescribeSubnetsOutput{
			Subnets: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsInstanceInvalidSubnetRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsInstanceInvalidVPCSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.SecurityGroup
		Expected tflint.Issues
	}{
		{
			Name: "security group is invalid",
			Content: `
resource "aws_instance" "web" {
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
					Rule:    NewAwsInstanceInvalidVpcSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 4, Column: 9},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
				{
					Rule:    NewAwsInstanceInvalidVpcSecurityGroupRule(),
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
resource "aws_instance" "web" {
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

resource "aws_instance" "web" {
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
					Rule:    NewAwsInstanceInvalidVpcSecurityGroupRule(),
					Message: "\"sg-1234abcd\" is invalid security group.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 30},
						End:      hcl.Pos{Line: 7, Column: 54},
					},
				},
				{
					Rule:    NewAwsInstanceInvalidVpcSecurityGroupRule(),
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

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidVPCSecurityGroup")
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
		rule := NewAwsInstanceInvalidVpcSecurityGroupRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsInstanceInvalidVpcSecurityGroupRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsLaunchConfigurationInvalidIAMProfile(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*iam.InstanceProfile
		Expected tflint.Issues
	}{
		{
			Name: "iam_instance_profile is invalid",
			Content: `
resource "aws_launch_configuration" "web" {
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
			Expected: tflint.Issues{
				{
					Rule:    NewAwsLaunchConfigurationInvalidIAMProfileRule(),
					Message: "\"app-server\" is invalid IAM profile name.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 40},
					},
				},
			},
		},
		{
			Name: "iam_instance_profile is valid",
			Content: `
resource "aws_launch_configuration" "web" {
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
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsLaunchConfigurationInvalidIamProfile")
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
		rule := NewAwsLaunchConfigurationInvalidIAMProfileRule()

		mock := client.NewMockIAMAPI(ctrl)
		mock.EXPECT().ListInstanceProfiles(&iam.ListInstanceProfilesInput{}).Return(&iam.ListInstanceProfilesOutput{
			InstanceProfiles: tc.Response,
		}, nil)
		runner.AwsClient.IAM = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsLaunchConfigurationInvalidIAMProfileRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidEgressOnlyGateway(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.EgressOnlyInternetGateway
		Expected tflint.Issues
	}{
		{
			Name: "egress only gateway id is invalid",
			Content: `
resource "aws_route" "foo" {
    egress_only_gateway_id = "igw-1234abcd"
}`,
			Response: []*ec2.EgressOnlyInternetGateway{
				{
					EgressOnlyInternetGatewayId: aws.String("eigw-1234abcd"),
				},
				{
					EgressOnlyInternetGatewayId: aws.String("eigw-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidEgressOnlyGatewayRule(),
					Message: "\"igw-1234abcd\" is invalid egress only internet gateway ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 30},
						End:      hcl.Pos{Line: 3, Column: 44},
					},
				},
			},
		},
		{
			Name: "egress only gateway id is valid",
			Content: `
resource "aws_route" "foo" {
    egress_only_gateway_id = "eigw-1234abcd"
}`,
			Response: []*ec2.EgressOnlyInternetGateway{
				{
					EgressOnlyInternetGatewayId: aws.String("eigw-1234abcd"),
				},
				{
					EgressOnlyInternetGatewayId: aws.String("eigw-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidEgressOnlyGateway")
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
		rule := NewAwsRouteInvalidEgressOnlyGatewayRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{}).Return(&ec2.DescribeEgressOnlyInternetGatewaysOutput{
			EgressOnlyInternetGateways: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidEgressOnlyGatewayRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidGateway(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.InternetGateway
		Expected tflint.Issues
	}{
		{
			Name: "gateway id is invalid",
			Content: `
resource "aws_route" "foo" {
    gateway_id = "eigw-1234abcd"
}`,
			Response: []*ec2.InternetGateway{
				{
					InternetGatewayId: aws.String("igw-1234abcd"),
				},
				{
					InternetGatewayId: aws.String("igw-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidGatewayRule(),
					Message: "\"eigw-1234abcd\" is invalid internet gateway ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 18},
						End:      hcl.Pos{Line: 3, Column: 33},
					},
				},
			},
		},
		{
			Name: "gateway id is valid",
			Content: `
resource "aws_route" "foo" {
    gateway_id = "igw-1234abcd"
}`,
			Response: []*ec2.InternetGateway{
				{
					InternetGatewayId: aws.String("igw-1234abcd"),
				},
				{
					InternetGatewayId: aws.String("igw-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidGateway")
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
		rule := NewAwsRouteInvalidGatewayRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{}).Return(&ec2.DescribeInternetGatewaysOutput{
			InternetGateways: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidGatewayRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidInstance(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.Instance
		Expected tflint.Issues
	}{
		{
			Name: "instance id is invalid",
			Content: `
resource "aws_route" "foo" {
    instance_id = "i-1234abcd"
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-5678abcd"),
				},
				{
					InstanceId: aws.String("i-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidInstanceRule(),
					Message: "\"i-1234abcd\" is invalid instance ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 19},
						End:      hcl.Pos{Line: 3, Column: 31},
					},
				},
			},
		},
		{
			Name: "instance id is valid",
			Content: `
resource "aws_route" "foo" {
    instance_id = "i-1234abcd"
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-1234abcd"),
				},
				{
					InstanceId: aws.String("i-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidInstance")
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
		rule := NewAwsRouteInvalidInstanceRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeInstances(&ec2.DescribeInstancesInput{}).Return(&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				{
					Instances: tc.Response,
				},
			},
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidInstanceRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidNatGateway(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.NatGateway
		Expected tflint.Issues
	}{
		{
			Name: "NAT gateway id is invalid",
			Content: `
resource "aws_route" "foo" {
    nat_gateway_id = "nat-1234abcd"
}`,
			Response: []*ec2.NatGateway{
				{
					NatGatewayId: aws.String("nat-5678abcd"),
				},
				{
					NatGatewayId: aws.String("nat-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidNatGatewayRule(),
					Message: "\"nat-1234abcd\" is invalid NAT gateway ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 22},
						End:      hcl.Pos{Line: 3, Column: 36},
					},
				},
			},
		},
		{
			Name: "NAT gateway id is valid",
			Content: `
resource "aws_route" "foo" {
    nat_gateway_id = "nat-1234abcd"
}`,
			Response: []*ec2.NatGateway{
				{
					NatGatewayId: aws.String("nat-1234abcd"),
				},
				{
					NatGatewayId: aws.String("nat-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidNatGateway")
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
		rule := NewAwsRouteInvalidNatGatewayRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeNatGateways(&ec2.DescribeNatGatewaysInput{}).Return(&ec2.DescribeNatGatewaysOutput{
			NatGateways: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidNatGatewayRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidNetworkInterface(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.NetworkInterface
		Expected tflint.Issues
	}{
		{
			Name: "network interface id is invalid",
			Content: `
resource "aws_route" "foo" {
    network_interface_id = "eni-1234abcd"
}`,
			Response: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-5678abcd"),
				},
				{
					NetworkInterfaceId: aws.String("eni-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidNetworkInterfaceRule(),
					Message: "\"eni-1234abcd\" is invalid network interface ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 42},
					},
				},
			},
		},
		{
			Name: "network interface id is valid",
			Content: `
resource "aws_route" "foo" {
    network_interface_id = "eni-1234abcd"
}`,
			Response: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-1234abcd"),
				},
				{
					NetworkInterfaceId: aws.String("eni-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidNetworkInterface")
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
		rule := NewAwsRouteInvalidNetworkInterfaceRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{}).Return(&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidNetworkInterfaceRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidRouteTable(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.RouteTable
		Expected tflint.Issues
	}{
		{
			Name: "route table id is invalid",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-nat-gw-a"
}`,
			Response: []*ec2.RouteTable{
				{
					RouteTableId: aws.String("rtb-1234abcd"),
				},
				{
					RouteTableId: aws.String("rtb-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidRouteTableRule(),
					Message: "\"rtb-nat-gw-a\" is invalid route table ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 22},
						End:      hcl.Pos{Line: 3, Column: 36},
					},
				},
			},
		},
		{
			Name: "route table id is valid",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
}`,
			Response: []*ec2.RouteTable{
				{
					RouteTableId: aws.String("rtb-1234abcd"),
				},
				{
					RouteTableId: aws.String("rtb-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidRouteTable")
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
		rule := NewAwsRouteInvalidRouteTableRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeRouteTables(&ec2.DescribeRouteTablesInput{}).Return(&ec2.DescribeRouteTablesOutput{
			RouteTables: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidRouteTableRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}

func Test_AwsRouteInvalidVPCPeeringConnection(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.VpcPeeringConnection
		Expected tflint.Issues
	}{
		{
			Name: "VPC peering connection id is invalid",
			Content: `
resource "aws_route" "foo" {
    vpc_peering_connection_id = "pcx-1234abcd"
}`,
			Response: []*ec2.VpcPeeringConnection{
				{
					VpcPeeringConnectionId: aws.String("pcx-5678abcd"),
				},
				{
					VpcPeeringConnectionId: aws.String("pcx-abcd1234"),
				},
			},
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteInvalidVpcPeeringConnectionRule(),
					Message: "\"pcx-1234abcd\" is invalid VPC peering connection ID.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 33},
						End:      hcl.Pos{Line: 3, Column: 47},
					},
				},
			},
		},
		{
			Name: "VPC peering connection id is valid",
			Content: `
resource "aws_route" "foo" {
    vpc_peering_connection_id = "pcx-1234abcd"
}`,
			Response: []*ec2.VpcPeeringConnection{
				{
					VpcPeeringConnectionId: aws.String("pcx-1234abcd"),
				},
				{
					VpcPeeringConnectionId: aws.String("pcx-abcd1234"),
				},
			},
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidVPCPeeringConnection")
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
		rule := NewAwsRouteInvalidVpcPeeringConnectionRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeVpcPeeringConnections(&ec2.DescribeVpcPeeringConnectionsInput{}).Return(&ec2.DescribeVpcPeeringConnectionsOutput{
			VpcPeeringConnections: tc.Response,
		}, nil)
		runner.AwsClient.EC2 = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteInvalidVpcPeeringConnectionRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}
