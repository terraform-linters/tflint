package tflint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_LoadConfig_v0_10_5(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "v0.10.5_module"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if _, exists := config.Children["ec2_instance"]; !exists {
		t.Fatalf("`ec2_instance` module is not loaded: %#v", config.Children)
	}

	if _, exists := config.Children["ec2_instance"].Module.ManagedResources["aws_instance.web"]; !exists {
		t.Fatalf("`ec2_instance` module resource `aws_instance.web` is not loaded: %#v", config.Children["ec2_instance"].Module.ManagedResources)
	}
}

func Test_LoadConfig_v0_10_6(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "v0.10.6_module"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if _, exists := config.Children["ec2_instance"]; !exists {
		t.Fatalf("`ec2_instance` module is not loaded: %#v", config.Children)
	}

	if _, exists := config.Children["ec2_instance"].Module.ManagedResources["aws_instance.web"]; !exists {
		t.Fatalf("`ec2_instance` module resource `aws_instance.web` is not loaded: %#v", config.Children["ec2_instance"].Module.ManagedResources)
	}
}

func Test_LoadConfig_v0_10_7(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "v0.10.7_module"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if _, exists := config.Children["ec2_instance"]; !exists {
		t.Fatalf("`ec2_instance` module is not loaded: %#v", config.Children)
	}

	if _, exists := config.Children["ec2_instance"].Module.ManagedResources["aws_instance.web"]; !exists {
		t.Fatalf("`ec2_instance` module resource `aws_instance.web` is not loaded: %#v", config.Children["ec2_instance"].Module.ManagedResources)
	}
}

func Test_LoadConfig_v0_11_0(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "v0.11.0_module"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	config, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if _, exists := config.Children["ecs_on_spotfleet"]; !exists {
		t.Fatalf("`ecs_on_spotfleet` module is not loaded: %#v", config.Children)
	}

	if _, exists := config.Children["ecs_on_spotfleet"].Module.ManagedResources["aws_ecs_cluster.main"]; !exists {
		t.Fatalf("`ecs_on_spotfleet` module resource `aws_ecs_cluster.main` is not loaded: %#v", config.Children["ecs_on_spotfleet"].Module.ManagedResources)
	}

	if _, exists := config.Children["consul"]; !exists {
		t.Fatalf("`consul` module is not loaded: %#v", config.Children)
	}

	if _, exists := config.Children["consul"].Children["consul_clients"]; !exists {
		t.Fatalf("`consul.consul_clients` module is not loaded: %#v", config.Children["consul"].Children)
	}

	if _, exists := config.Children["consul"].Children["consul_clients"].Children["iam_policies"]; !exists {
		t.Fatalf("`consule.consul_clients.iam_policies` module is not loaded: %#v", config.Children["consul"].Children["consul_clients"].Children)
	}

	if _, exists := config.Children["consul"].Children["consul_clients"].Children["iam_policies"].Module.ManagedResources["aws_iam_role_policy.auto_discover_cluster"]; !exists {
		t.Fatalf("`consule.consul_clients.iam_policies` module resource `aws_iam_role_policy.auto_discover_cluster` is not loaded: %#v", config.Children["consul"].Children["consul_clients"].Children["iam_policies"].Module.ManagedResources)
	}

	if _, exists := config.Children["consul"].Children["consul_clients"].Children["security_group_rules"]; !exists {
		t.Fatalf("`consule.consul_clients.security_group_rules` module is not loaded: %#v", config.Children["consul"].Children["consul_clients"].Children)
	}

	if _, exists := config.Children["consul"].Children["consul_clients"].Children["security_group_rules"].Module.ManagedResources["aws_security_group_rule.allow_server_rpc_inbound"]; !exists {
		t.Fatalf("`consule.consul_clients.security_group_rules` module resource `aws_security_group_rule.allow_server_rpc_inbound` is not loaded: %#v", config.Children["consul"].Children["consul_clients"].Children["security_group_rules"].Module.ManagedResources)
	}

	if _, exists := config.Children["consul"].Children["consul_servers"]; !exists {
		t.Fatalf("`consul.consul_servers` module is not loaded: %#v", config.Children["consul"].Children)
	}

	if _, exists := config.Children["consul"].Children["consul_servers"].Children["iam_policies"]; !exists {
		t.Fatalf("`consule.consul_servers.iam_policies` module is not loaded: %#v", config.Children["consul"].Children["consul_servers"].Children)
	}

	if _, exists := config.Children["consul"].Children["consul_servers"].Children["iam_policies"].Module.ManagedResources["aws_iam_role_policy.auto_discover_cluster"]; !exists {
		t.Fatalf("`consule.consul_servers.iam_policies` module resource `aws_iam_role_policy.auto_discover_cluster` is not loaded: %#v", config.Children["consul"].Children["consul_servers"].Children["iam_policies"].Module.ManagedResources)
	}

	if _, exists := config.Children["consul"].Children["consul_servers"].Children["security_group_rules"]; !exists {
		t.Fatalf("`consule.consul_servers.security_group_rules` module is not loaded: %#v", config.Children["consul"].Children["consul_servers"].Children)
	}

	if _, exists := config.Children["consul"].Children["consul_servers"].Children["security_group_rules"].Module.ManagedResources["aws_security_group_rule.allow_server_rpc_inbound"]; !exists {
		t.Fatalf("`consule.consul_servers.security_group_rules` module resource `aws_security_group_rule.allow_server_rpc_inbound` is not loaded: %#v", config.Children["consul"].Children["consul_servers"].Children["security_group_rules"].Module.ManagedResources)
	}
}

func Test_LoadConfig_moduleNotFound(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "before_terraform_init"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	_, err = loader.LoadConfig()
	if err == nil {
		t.Fatal("Expected error is not occurred")
	}

	expected := "module.tf:1,1-22: `ec2_instance` module is not found. Did you run `terraform init`?; "
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func Test_LoadConfig_invalidConfiguration(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "invalid_configuration"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	_, err = loader.LoadConfig()
	if err == nil {
		t.Fatal("Expected error is not occurred")
	}

	expected := "resource.tf:1,1-10: Unsupported block type; Blocks of type \"resources\" are not expected here. Did you mean \"resource\"?"
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func Test_LoadValuesFiles(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "values_files"))
	if err != nil {
		t.Fatal(err)
	}
	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	ret, err := loader.LoadValuesFiles("cli1.tfvars", "cli2.tfvars")
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := []terraform.InputValues{
		{
			"default": {
				Value:      cty.StringVal("terraform.tfvars"),
				SourceType: terraform.ValueFromAutoFile,
			},
		},
		{
			"auto1": {
				Value:      cty.StringVal("auto1.auto.tfvars"),
				SourceType: terraform.ValueFromAutoFile,
			},
		},
		{
			"auto2": {
				Value:      cty.StringVal("auto2.auto.tfvars"),
				SourceType: terraform.ValueFromAutoFile,
			},
		},
		{
			"cli1": {
				Value:      cty.StringVal("cli1.tfvars"),
				SourceType: terraform.ValueFromNamedFile,
			},
		},
		{
			"cli2": {
				Value:      cty.StringVal("cli2.tfvars"),
				SourceType: terraform.ValueFromNamedFile,
			},
		},
	}

	if !reflect.DeepEqual(expected, ret) {
		t.Fatalf("Unexpected input values are received: expected=%#v actual=%#v", expected, ret)
	}
}

func Test_LoadValuesFiles_invalidValuesFile(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "invalid_values_files"))
	if err != nil {
		t.Fatal(err)
	}
	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	_, err = loader.LoadValuesFiles()
	if err == nil {
		t.Fatal("Expected error is not occurred")
	}

	expected := "terraform.tfvars: Unexpected resource block; Blocks are not allowed here."
	if err.Error() != expected {
		t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
	}
}

func Test_IsConfigFile(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "is_config_file"))
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	cases := []struct {
		Name     string
		File     string
		Expected bool
	}{
		{
			Name:     "config file",
			File:     "template.tf",
			Expected: true,
		},
		{
			Name:     "not config file",
			File:     "README",
			Expected: false,
		},
		{
			Name:     "not found",
			File:     "not_found.tf",
			Expected: false,
		},
	}

	for _, tc := range cases {
		ret := loader.IsConfigFile(tc.File)
		if ret != tc.Expected {
			t.Fatalf("Test `%s` failed: expected is `%t`, but get `%t`", tc.Name, tc.Expected, ret)
		}
	}
}
