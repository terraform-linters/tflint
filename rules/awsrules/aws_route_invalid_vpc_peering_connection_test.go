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

func Test_AwsRouteInvalidVPCPeeringConnection(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.VpcPeeringConnection
		Expected issue.Issues
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
			Expected: []*issue.Issue{
				{
					Detector: "aws_route_invalid_vpc_peering_connection",
					Type:     "ERROR",
					Message:  "\"pcx-1234abcd\" is invalid VPC peering connection ID.",
					Line:     3,
					File:     "resource.tf",
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
			Expected: []*issue.Issue{},
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
		rule := NewAwsRouteInvalidVPCPeeringConnectionRule()

		mock := client.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeVpcPeeringConnections(&ec2.DescribeVpcPeeringConnectionsInput{}).Return(&ec2.DescribeVpcPeeringConnectionsOutput{
			VpcPeeringConnections: tc.Response,
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
