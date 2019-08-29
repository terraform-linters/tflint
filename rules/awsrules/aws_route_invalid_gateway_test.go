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
