package terraform

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

func TestLoadConfig_v0_15_0(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		config, diags := loader.LoadConfig(".", true)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		// root
		if config.Module.SourceDir != "." {
			t.Fatalf("root module path: want=%s, got=%s", ".", config.Module.SourceDir)
		}
		// module.instance
		testChildModule(t, config, "instance", "ec2")
		// module.consul
		testChildModule(t, config, "consul", filepath.Join(".terraform", "modules", "consul"))
		// module.consul.module.consul_clients
		testChildModule(
			t,
			config.Children["consul"],
			"consul_clients",
			filepath.Join(".terraform", "modules", "consul", "modules", "consul-cluster"),
		)
		// module.consul.module.consul_clients.module.iam_policies
		testChildModule(
			t,
			config.Children["consul"].Children["consul_clients"],
			"iam_policies",
			filepath.Join(".terraform", "modules", "consul", "modules", "consul-iam-policies"),
		)
		// module.consul.module.consul_clients.module.security_group_rules
		testChildModule(
			t,
			config.Children["consul"].Children["consul_clients"],
			"security_group_rules",
			filepath.Join(".terraform", "modules", "consul", "modules", "consul-security-group-rules"),
		)
		// module.consul.module.consul_servers
		testChildModule(
			t,
			config.Children["consul"],
			"consul_servers",
			filepath.Join(".terraform", "modules", "consul", "modules", "consul-cluster"),
		)
		// module.consul.module.consul_servers.module.iam_policies
		testChildModule(
			t,
			config.Children["consul"].Children["consul_servers"],
			"iam_policies",
			filepath.Join(".terraform", "modules", "consul", "modules", "consul-iam-policies"),
		)
		// module.consul.module.consul_servers.module.security_group_rules
		testChildModule(
			t,
			config.Children["consul"].Children["consul_servers"],
			"security_group_rules",
			filepath.Join(".terraform", "modules", "consul", "modules", "consul-security-group-rules"),
		)
	})
}

func TestLoadConfig_v0_15_0_withBaseDir(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func(dir string) {
		// The current dir is test-fixtures/v0.15.0_module, but the base dir is test-fixtures
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, filepath.Dir(dir))
		if err != nil {
			t.Fatal(err)
		}
		config, diags := loader.LoadConfig(".", true)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		// root
		if config.Module.SourceDir != "v0.15.0_module" {
			t.Fatalf("root module path: want=%s, got=%s", "v0.15.0_module", config.Module.SourceDir)
		}
		// module.instance
		testChildModule(t, config, "instance", filepath.Join("v0.15.0_module", "ec2"))
		// module.consul
		testChildModule(t, config, "consul", filepath.Join("v0.15.0_module", ".terraform", "modules", "consul"))
		// module.consul.module.consul_clients
		testChildModule(
			t,
			config.Children["consul"],
			"consul_clients",
			filepath.Join("v0.15.0_module", ".terraform", "modules", "consul", "modules", "consul-cluster"),
		)
		// module.consul.module.consul_clients.module.iam_policies
		testChildModule(
			t,
			config.Children["consul"].Children["consul_clients"],
			"iam_policies",
			filepath.Join("v0.15.0_module", ".terraform", "modules", "consul", "modules", "consul-iam-policies"),
		)
		// module.consul.module.consul_clients.module.security_group_rules
		testChildModule(
			t,
			config.Children["consul"].Children["consul_clients"],
			"security_group_rules",
			filepath.Join("v0.15.0_module", ".terraform", "modules", "consul", "modules", "consul-security-group-rules"),
		)
		// module.consul.module.consul_servers
		testChildModule(
			t,
			config.Children["consul"],
			"consul_servers",
			filepath.Join("v0.15.0_module", ".terraform", "modules", "consul", "modules", "consul-cluster"),
		)
		// module.consul.module.consul_servers.module.iam_policies
		testChildModule(
			t,
			config.Children["consul"].Children["consul_servers"],
			"iam_policies",
			filepath.Join("v0.15.0_module", ".terraform", "modules", "consul", "modules", "consul-iam-policies"),
		)
		// module.consul.module.consul_servers.module.security_group_rules
		testChildModule(
			t,
			config.Children["consul"].Children["consul_servers"],
			"security_group_rules",
			filepath.Join("v0.15.0_module", ".terraform", "modules", "consul", "modules", "consul-security-group-rules"),
		)
	})
}

func TestLoadConfig_moduleNotFound(t *testing.T) {
	withinFixtureDir(t, "before_terraform_init", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		_, diags := loader.LoadConfig(".", true)
		if !diags.HasErrors() {
			t.Fatal("Expected error is not occurred")
		}

		expected := "module.tf:1,1-22: `ec2_instance` module is not found. Did you run `terraform init`?; "
		if diags.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, diags)
		}
	})
}

func TestLoadConfig_disableModules(t *testing.T) {
	withinFixtureDir(t, "before_terraform_init", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		config, diags := loader.LoadConfig(".", false)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		if config.Module.SourceDir != "." {
			t.Fatalf("Root module path: want=%s, got=%s", ".", config.Module.SourceDir)
		}
		if len(config.Children) != 0 {
			t.Fatalf("Root module has children unexpectedly: %#v", config.Children)
		}
	})
}

func TestLoadConfig_disableModules_withArgDir(t *testing.T) {
	withinFixtureDir(t, ".", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		config, diags := loader.LoadConfig("before_terraform_init", false)
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		if config.Module.SourceDir != "before_terraform_init" {
			t.Fatalf("Root module path: want=%s, got=%s", "before_terraform_init", config.Module.SourceDir)
		}
		if len(config.Children) != 0 {
			t.Fatalf("Root module has children unexpectedly: %#v", config.Children)
		}
	})
}

func TestLoadConfig_invalidConfiguration(t *testing.T) {
	withinFixtureDir(t, "invalid_configuration", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		_, diags := loader.LoadConfig(".", false)
		if !diags.HasErrors() {
			t.Fatal("Expected error is not occurred")
		}

		expected := "resource.tf:3,23-29: Missing newline after argument; An argument definition must end with a newline."
		if diags.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, diags)
		}
	})
}

func TestLoadValuesFiles(t *testing.T) {
	withinFixtureDir(t, "values_files", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
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

		want := []string{
			"auto1.auto.tfvars",
			"auto2.auto.tfvars",
			"cli1.tfvars",
			"cli2.tfvars",
			"terraform.tfvars",
		}
		loadedFiles := []string{}
		for name := range loader.Files() {
			loadedFiles = append(loadedFiles, name)
		}
		opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })
		if diff := cmp.Diff(want, loadedFiles, opt); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestLoadValuesFiles_withBaseDir(t *testing.T) {
	withinFixtureDir(t, "values_files", func(dir string) {
		// The current dir is test-fixtures/values_files, but the base dir is test-fixtures
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, filepath.Dir(dir))
		if err != nil {
			t.Fatal(err)
		}
		// Files passed manually are relative to the current directory.
		ret, diags := loader.LoadValuesFiles(
			".",
			"cli1.tfvars",
			"cli2.tfvars",
		)
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

		want := []string{
			filepath.Join("values_files", "auto1.auto.tfvars"),
			filepath.Join("values_files", "auto2.auto.tfvars"),
			filepath.Join("values_files", "cli1.tfvars"),
			filepath.Join("values_files", "cli2.tfvars"),
			filepath.Join("values_files", "terraform.tfvars"),
		}
		loadedFiles := []string{}
		for name := range loader.Files() {
			loadedFiles = append(loadedFiles, name)
		}
		opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })
		if diff := cmp.Diff(want, loadedFiles, opt); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestLoadValuesFiles_withArgDir(t *testing.T) {
	withinFixtureDir(t, ".", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		ret, diags := loader.LoadValuesFiles(
			"values_files",
			filepath.Join("values_files", "cli1.tfvars"),
			filepath.Join("values_files", "cli2.tfvars"),
		)
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

		want := []string{
			filepath.Join("values_files", "auto1.auto.tfvars"),
			filepath.Join("values_files", "auto2.auto.tfvars"),
			filepath.Join("values_files", "cli1.tfvars"),
			filepath.Join("values_files", "cli2.tfvars"),
			filepath.Join("values_files", "terraform.tfvars"),
		}
		loadedFiles := []string{}
		for name := range loader.Files() {
			loadedFiles = append(loadedFiles, name)
		}
		opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })
		if diff := cmp.Diff(want, loadedFiles, opt); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestLoadValuesFiles_invalidValuesFile(t *testing.T) {
	withinFixtureDir(t, "invalid_values_files", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		_, diags := loader.LoadValuesFiles(".")
		if !diags.HasErrors() {
			t.Fatal("Expected error is not occurred")
		}

		expected := "terraform.tfvars:3,1-9: Unexpected \"resource\" block; Blocks are not allowed here."
		if diags.Error() != expected {
			t.Fatalf("Expected error is `%s`, but get `%s`", expected, diags)
		}
	})
}

func TestLoadConfigDirFiles_v0_15_0(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		files, diags := loader.LoadConfigDirFiles(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		want := []string{"module.tf"}
		loadedFiles := []string{}
		for name := range files {
			loadedFiles = append(loadedFiles, name)
		}
		opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })
		if diff := cmp.Diff(want, loadedFiles, opt); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestLoadConfigDirFiles_v0_15_0_withBaseDir(t *testing.T) {
	withinFixtureDir(t, "v0.15.0_module", func(dir string) {
		// The current dir is test-fixtures/v0.15.0_module, but the base dir is test-fixtures
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, filepath.Dir(dir))
		if err != nil {
			t.Fatal(err)
		}
		files, diags := loader.LoadConfigDirFiles(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		want := []string{filepath.Join("v0.15.0_module", "module.tf")}
		loadedFiles := []string{}
		for name := range files {
			loadedFiles = append(loadedFiles, name)
		}
		opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })
		if diff := cmp.Diff(want, loadedFiles, opt); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestLoadConfigDirFiles_v0_15_0_withArgDir(t *testing.T) {
	withinFixtureDir(t, ".", func(dir string) {
		loader, err := NewLoader(afero.Afero{Fs: afero.NewOsFs()}, dir)
		if err != nil {
			t.Fatal(err)
		}
		files, diags := loader.LoadConfigDirFiles("v0.15.0_module")
		if diags.HasErrors() {
			t.Fatal(diags)
		}

		want := []string{filepath.Join("v0.15.0_module", "module.tf")}
		loadedFiles := []string{}
		for name := range files {
			loadedFiles = append(loadedFiles, name)
		}
		opt := cmpopts.SortSlices(func(x, y string) bool { return x > y })
		if diff := cmp.Diff(want, loadedFiles, opt); diff != "" {
			t.Fatal(diff)
		}
	})
}

func withinFixtureDir(t *testing.T, dir string, test func(string)) {
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

	workingDir := filepath.Join(currentDir, "test-fixtures", dir)
	if err = os.Chdir(workingDir); err != nil {
		t.Fatal(err)
	}

	test(workingDir)
}

func testChildModule(t *testing.T, config *Config, key string, wantPath string) {
	t.Helper()

	if _, exists := config.Children[key]; !exists {
		t.Fatalf("`%s` module is not loaded: %#v", key, config.Children)
	}
	modulePath := config.Children[key].Module.SourceDir
	if modulePath != wantPath {
		t.Fatalf("`%s` module path: want=%s, got=%s", key, wantPath, modulePath)
	}
}
