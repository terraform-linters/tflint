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
