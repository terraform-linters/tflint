package terraform

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

func TestLoadConfig_v0_15_0(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()})
		if err != nil {
			t.Fatal(err)
		}
		config, diags := loader.LoadConfig(".", true)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		if _, exists := config.Children["instance"]; !exists {
			t.Fatalf("`instance` module is not loaded: %#v", config.Children)
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

		if _, exists := config.Children["consul"].Children["consul_clients"].Children["security_group_rules"]; !exists {
			t.Fatalf("`consule.consul_clients.security_group_rules` module is not loaded: %#v", config.Children["consul"].Children["consul_clients"].Children)
		}

		if _, exists := config.Children["consul"].Children["consul_servers"]; !exists {
			t.Fatalf("`consul.consul_servers` module is not loaded: %#v", config.Children["consul"].Children)
		}

		if _, exists := config.Children["consul"].Children["consul_servers"].Children["iam_policies"]; !exists {
			t.Fatalf("`consule.consul_servers.iam_policies` module is not loaded: %#v", config.Children["consul"].Children["consul_servers"].Children)
		}

		if _, exists := config.Children["consul"].Children["consul_servers"].Children["security_group_rules"]; !exists {
			t.Fatalf("`consule.consul_servers.security_group_rules` module is not loaded: %#v", config.Children["consul"].Children["consul_servers"].Children)
		}
	})
}

func TestLoadConfig_moduleNotFound(t *testing.T) {
	withinFixtureDir(t, "before_terraform_init", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()})
		if err != nil {
			t.Fatal(err)
		}
		_, diags := loader.LoadConfig(".", true)
		if !diags.HasErrors() {
			t.Fatal("Expected error is not occurred")
		}

		expected := "module.tf:1,1-22: `ec2_instance` module is not found. Did you run `terraform init`?; "
		if diags.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
		}
	})
}

func TestLoadConfig_disableModules(t *testing.T) {
	withinFixtureDir(t, "before_terraform_init", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()})
		if err != nil {
			t.Fatal(err)
		}
		config, diags := loader.LoadConfig(".", false)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		if len(config.Children) != 0 {
			t.Fatalf("Root module has children unexpectedly: %#v", config.Children)
		}
	})
}

func TestLoadConfig_invalidConfiguration(t *testing.T) {
	withinFixtureDir(t, "invalid_configuration", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()})
		if err != nil {
			t.Fatal(err)
		}
		_, diags := loader.LoadConfig(".", false)
		if !diags.HasErrors() {
			t.Fatal("Expected error is not occurred")
		}

		expected := "resource.tf:3,23-29: Missing newline after argument; An argument definition must end with a newline."
		if diags.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
		}
	})
}

func Test_LoadValuesFiles(t *testing.T) {
	withinFixtureDir(t, "values_files", func() {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()})
		if err != nil {
			t.Fatal(err)
		}
		ret, diags := loader.LoadValuesFiles(".", "cli1.tfvars", "cli2.tfvars")
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		expected := []InputValues{
			{
				"default": {
					Value: cty.StringVal("terraform.tfvars"),
				},
			},
			{
				"auto1": {
					Value: cty.StringVal("auto1.auto.tfvars"),
				},
			},
			{
				"auto2": {
					Value: cty.StringVal("auto2.auto.tfvars"),
				},
			},
			{
				"cli1": {
					Value: cty.StringVal("cli1.tfvars"),
				},
			},
			{
				"cli2": {
					Value: cty.StringVal("cli2.tfvars"),
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
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()})
		if err != nil {
			t.Fatal(err)
		}
		_, diags := loader.LoadValuesFiles(".")
		if !diags.HasErrors() {
			t.Fatal("Expected error is not occurred")
		}

		expected := "terraform.tfvars:3,1-9: Unexpected \"resource\" block; Blocks are not allowed here."
		if diags.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, err.Error())
		}
	})
}

func withinFixtureDir(t *testing.T, dir string, test func()) {
	t.Helper()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(currentDir); err != nil {
			t.Fatal(err)
		}
	}()

	if err = os.Chdir(filepath.Join(currentDir, "test-fixtures", dir)); err != nil {
		t.Fatal(err)
	}

	test()
}
