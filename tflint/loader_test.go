package tflint

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_LoadConfig_v0_15_0(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, moduleConfig())
		if err != nil {
			t.Fatal(err)
		}
		config, err := loader.LoadConfig(".")
		if err != nil {
			t.Fatal(err)
		}

		if _, exists := config.Children["instance"]; !exists {
			t.Fatalf("`instance` module is not loaded: %#v", config.Children)
		}

		if _, exists := config.Children["instance"].Module.ManagedResources["aws_instance.main"]; !exists {
			t.Fatalf("`instance` module resource `aws_instance.main` is not loaded: %#v", config.Children["instance"].Module.ManagedResources)
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
	})
}

func Test_LoadConfig_moduleNotFound(t *testing.T) {
	withinFixtureDir(t, "before_terraform_init", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, moduleConfig())
		if err != nil {
			t.Fatal(err)
		}
		_, err = loader.LoadConfig(".")
		if err == nil {
			t.Fatal("Expected error is not occurred")
		}

		expected := "module.tf:1,1-22: `ec2_instance` module is not found. Did you run `terraform init`?; "
		if err.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
		}
	})
}

func Test_LoadConfig_disableModules(t *testing.T) {
	withinFixtureDir(t, "before_terraform_init", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, EmptyConfig())
		if err != nil {
			t.Fatal(err)
		}
		config, err := loader.LoadConfig(".")
		if err != nil {
			t.Fatal(err)
		}

		if len(config.Children) != 0 {
			t.Fatalf("Root module has children unexpectedly: %#v", config.Children)
		}
	})
}

func Test_LoadConfig_invalidConfiguration(t *testing.T) {
	withinFixtureDir(t, "invalid_configuration", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, EmptyConfig())
		if err != nil {
			t.Fatal(err)
		}
		_, err = loader.LoadConfig(".")
		if err == nil {
			t.Fatal("Expected error is not occurred")
		}

		expected := "resource.tf:1,1-10: Unsupported block type; Blocks of type \"resources\" are not expected here. Did you mean \"resource\"?"
		if err.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
		}
	})
}

func Test_Files(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, EmptyConfig())
		if err != nil {
			t.Fatal(err)
		}
		_, err = loader.LoadConfig(".")
		if err != nil {
			t.Fatal(err)
		}

		files, err := loader.Files()
		if err != nil {
			t.Fatal(err)
		}

		filename := "module.tf"
		b, err := afero.ReadFile(loader.fs, filename)
		if err != nil {
			t.Fatal(err)
		}

		expected := map[string]*hcl.File{
			"module.tf": {Bytes: b},
		}

		opts := cmpopts.IgnoreFields(hcl.File{}, "Body", "Nav")
		if !cmp.Equal(expected, files, opts) {
			t.Fatalf("Test failed. Diff: %s", cmp.Diff(expected, files, opts))
		}
	})
}

func Test_LoadAnnotations(t *testing.T) {
	withinFixtureDir(t, "annotation_files", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, EmptyConfig())
		if err != nil {
			t.Fatal(err)
		}
		ret, err := loader.LoadAnnotations(".")
		if err != nil {
			t.Fatal(err)
		}

		expected := map[string]Annotations{
			"file1.tf": {
				{
					Content: "aws_instance_invalid_type",
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenComment,
						Bytes: []byte(fmt.Sprintf("// tflint-ignore: aws_instance_invalid_type%s", newLine())),
						Range: hcl.Range{
							Filename: "file1.tf",
							Start:    hcl.Pos{Line: 2, Column: 5},
							End:      hcl.Pos{Line: 3, Column: 1},
						},
					},
				},
			},
			"file2.tf": {
				{
					Content: "aws_instance_invalid_type",
					Token: hclsyntax.Token{
						Type:  hclsyntax.TokenComment,
						Bytes: []byte(fmt.Sprintf("// tflint-ignore: aws_instance_invalid_type%s", newLine())),
						Range: hcl.Range{
							Filename: "file2.tf",
							Start:    hcl.Pos{Line: 2, Column: 32},
							End:      hcl.Pos{Line: 3, Column: 1},
						},
					},
				},
			},
			"file3.tf": {},
		}

		opts := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
		if !cmp.Equal(expected, ret, opts) {
			t.Fatalf("Test failed. Diff: %s", cmp.Diff(expected, ret, opts))
		}
	})
}

func Test_LoadValuesFiles(t *testing.T) {
	withinFixtureDir(t, "values_files", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, EmptyConfig())
		if err != nil {
			t.Fatal(err)
		}
		ret, err := loader.LoadValuesFiles("cli1.tfvars", "cli2.tfvars")
		if err != nil {
			t.Fatal(err)
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
	})
}

func Test_LoadValuesFiles_invalidValuesFile(t *testing.T) {
	withinFixtureDir(t, "invalid_values_files", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, EmptyConfig())
		if err != nil {
			t.Fatal(err)
		}
		_, err = loader.LoadValuesFiles()
		if err == nil {
			t.Fatal("Expected error is not occurred")
		}

		expected := "terraform.tfvars:3,1-9: Unexpected \"resource\" block; Blocks are not allowed here."
		if err.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
		}
	})
}
