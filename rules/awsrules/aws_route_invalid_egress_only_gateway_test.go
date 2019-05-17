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
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsRouteInvalidEgressOnlyGateway(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*ec2.EgressOnlyInternetGateway
		Expected issue.Issues
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
			Expected: []*issue.Issue{
				{
					Detector: "aws_route_invalid_egress_only_gateway",
					Type:     "ERROR",
					Message:  "\"igw-1234abcd\" is invalid egress only internet gateway ID.",
					Line:     3,
					File:     "resource.tf",
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
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteInvalidEgressOnlyGateway")
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
		rule := NewAwsRouteInvalidEgressOnlyGatewayRule()

		mock := mock.NewMockEC2API(ctrl)
		mock.EXPECT().DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{}).Return(&ec2.DescribeEgressOnlyInternetGatewaysOutput{
			EgressOnlyInternetGateways: tc.Response,
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
