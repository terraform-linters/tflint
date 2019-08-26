package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsRouteNotSpecifiedTarget(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "route target is not specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteNotSpecifiedTargetRule(),
					Message: "The routing target is not specified, each aws_route must contain either egress_only_gateway_id, gateway_id, instance_id, nat_gateway_id, network_interface_id, transit_gateway_id, or vpc_peering_connection_id.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 27},
					},
				},
			},
		},
		{
			Name: "gateway_id is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "egress_only_gateway_id is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    egress_only_gateway_id = "eigw-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "nat_gateway_id is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    nat_gateway_id = "nat-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "instance_id is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    instance_id = "i-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "vpc_peering_connection_id is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    vpc_peering_connection_id = "pcx-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "network_interface_id is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    network_interface_id = "eni-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "transit_gateway_id is specified",
			Content: `
resource "aws_route" "foo" {
	route_table_id = "rtb-1234abcd"
	transit_gateway_id = "tgw-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "transit_gateway_id is specified, but the value is null",
			Content: `
variable "transit_gateway_id" {
	type    = string
	default = null
}

resource "aws_route" "foo" {
	route_table_id = "rtb-1234abcd"
	transit_gateway_id = var.transit_gateway_id
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteNotSpecifiedTargetRule(),
					Message: "The routing target is not specified, each aws_route must contain either egress_only_gateway_id, gateway_id, instance_id, nat_gateway_id, network_interface_id, transit_gateway_id, or vpc_peering_connection_id.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 7, Column: 1},
						End:      hcl.Pos{Line: 7, Column: 27},
					},
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsRouteNotSpecifiedTarget")
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
		rule := NewAwsRouteNotSpecifiedTargetRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsRouteNotSpecifiedTargetRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}
