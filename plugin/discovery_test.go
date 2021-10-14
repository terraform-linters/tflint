package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_Discovery(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugins")
	defer func() { PluginRoot = original }()

	plugin, err := Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
			},
			"bar": {
				Name:    "bar",
				Enabled: false,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			},
		},
	})
	defer plugin.Clean()

	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}

	if len(plugin.RuleSets) != 1 {
		t.Fatalf("Only one plugin must be enabled, but %d plugins are enabled", len(plugin.RuleSets))
	}
}

func Test_Discovery_local(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	err = os.Chdir(filepath.Join(cwd, "test-fixtures", "locals"))
	if err != nil {
		t.Fatal(err)
	}

	plugin, err := Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: false,
			},
			"bar": {
				Name:    "bar",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			},
		},
	})
	defer plugin.Clean()

	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}

	if len(plugin.RuleSets) != 1 {
		t.Fatalf("Only one plugin must be enabled, but %d plugins are enabled", len(plugin.RuleSets))
	}
}

func Test_Discovery_envVar(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("TFLINT_PLUGIN_DIR", filepath.Join(cwd, "test-fixtures", "locals", ".tflint.d", "plugins"))
	defer os.Setenv("TFLINT_PLUGIN_DIR", "")

	plugin, err := Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
			},
			"bar": {
				Name:    "bar",
				Enabled: false,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			},
		},
	})
	defer plugin.Clean()

	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}

	if len(plugin.RuleSets) != 1 {
		t.Fatalf("Only one plugin must be enabled, but %d plugins are enabled", len(plugin.RuleSets))
	}
}

func Test_Discovery_pluginDirConfig(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	plugin, err := Discovery(&tflint.Config{
		PluginDir: filepath.Join(cwd, "test-fixtures", "locals", ".tflint.d", "plugins"),
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
			},
			"bar": {
				Name:    "bar",
				Enabled: false,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			},
		},
	})
	defer plugin.Clean()

	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}

	if len(plugin.RuleSets) != 1 {
		t.Fatalf("Only one plugin must be enabled, but %d plugins are enabled", len(plugin.RuleSets))
	}
}

func Test_Discovery_notFound(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	err = os.Chdir(filepath.Join(cwd, "test-fixtures", "no_plugins"))
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "no_plugins")
	defer func() { PluginRoot = original }()

	_, err = Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
			},
		},
	})

	if err == nil {
		t.Fatal("The error should have occurred, but didn't")
	}
	expected := fmt.Sprintf("Plugin `foo` not found in %s", PluginRoot)
	if err.Error() != expected {
		t.Fatalf("The error message is not matched: want=%s, got=%s", expected, err.Error())
	}
}

func Test_Discovery_plugin_name_is_directory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	err = os.Chdir(filepath.Join(cwd, "test-fixtures", "plugin_name_is_directory"))
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugin_name_is_directory")
	defer func() { PluginRoot = original }()

	_, err = Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
			},
		},
	})

	if err == nil {
		t.Fatal("The error should have occurred, but didn't")
	}
	expected := fmt.Sprintf("Plugin `foo` not found in %s", PluginRoot)
	if err.Error() != expected {
		t.Fatalf("The error message is not matched: want=%s, got=%s", expected, err.Error())
	}
}

func Test_Discovery_notFoundForAutoInstallation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	err = os.Chdir(filepath.Join(cwd, "test-fixtures", "no_plugins"))
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "no_plugins")
	defer func() { PluginRoot = original }()

	_, err = Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-foo",
				Version: "0.1.0",
			},
		},
	})

	if err == nil {
		t.Fatal("An error should have occurred, but it did not occur")
	}
	expected := "Plugin `foo` not found. Did you run `tflint --init`?"
	if err.Error() != expected {
		t.Fatalf("Error message not matched: want=%s, got=%s", expected, err.Error())
	}
}

func Test_FindPluginPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugins")
	defer func() { PluginRoot = original }()

	cases := []struct {
		Name     string
		Input    *InstallConfig
		Expected string
	}{
		{
			Name:     "manually installed",
			Input:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{Name: "foo", Enabled: true}),
			Expected: filepath.Join(PluginRoot, "tflint-ruleset-foo"+fileExt()),
		},
		{
			Name: "auto installed",
			Input: NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
				Name:    "bar",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			}),
			Expected: filepath.Join(PluginRoot, "github.com/terraform-linters/tflint-ruleset-bar", "0.1.0", "tflint-ruleset-bar"+fileExt()),
		},
	}

	for _, tc := range cases {
		got, err := FindPluginPath(tc.Input)
		if err != nil {
			t.Fatalf("Unexpected error occurred %s", err)
		}
		if got != tc.Expected {
			t.Fatalf("Failed `%s`: want=%s got=%s", tc.Name, tc.Expected, got)
		}
	}
}

func Test_FindPluginPath_locals(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	dir := filepath.Join(cwd, "test-fixtures", "locals")
	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Input    *InstallConfig
		Expected string
	}{
		{
			Name:     "manually installed",
			Input:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{Name: "foo", Enabled: true}),
			Expected: filepath.Join(localPluginRoot, "tflint-ruleset-foo"+fileExt()),
		},
		{
			Name: "auto installed",
			Input: NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
				Name:    "bar",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			}),
			Expected: filepath.Join(localPluginRoot, "github.com/terraform-linters/tflint-ruleset-bar", "0.1.0", "tflint-ruleset-bar"+fileExt()),
		},
	}

	for _, tc := range cases {
		got, err := FindPluginPath(tc.Input)
		if err != nil {
			t.Fatalf("Unexpected error occurred %s", err)
		}
		if got != tc.Expected {
			t.Fatalf("Failed `%s`: want=%s got=%s", tc.Name, tc.Expected, got)
		}
	}
}

func Test_FindPluginPath_envVar(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(cwd, "test-fixtures", "locals", ".tflint.d", "plugins")
	os.Setenv("TFLINT_PLUGIN_DIR", dir)
	defer os.Setenv("TFLINT_PLUGIN_DIR", "")

	cases := []struct {
		Name     string
		Input    *InstallConfig
		Expected string
	}{
		{
			Name:     "manually installed",
			Input:    NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{Name: "foo", Enabled: true}),
			Expected: filepath.Join(dir, "tflint-ruleset-foo"+fileExt()),
		},
		{
			Name: "auto installed",
			Input: NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
				Name:    "bar",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			}),
			Expected: filepath.Join(dir, "github.com/terraform-linters/tflint-ruleset-bar", "0.1.0", "tflint-ruleset-bar"+fileExt()),
		},
	}

	for _, tc := range cases {
		got, err := FindPluginPath(tc.Input)
		if err != nil {
			t.Fatalf("Unexpected error occurred %s", err)
		}
		if got != tc.Expected {
			t.Fatalf("Failed `%s`: want=%s got=%s", tc.Name, tc.Expected, got)
		}
	}
}

func Test_FindPluginPath_pluginDirConfig(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	globalConfig := tflint.EmptyConfig()
	globalConfig.PluginDir = filepath.Join(cwd, "test-fixtures", "locals", ".tflint.d", "plugins")

	cases := []struct {
		Name     string
		Input    *InstallConfig
		Expected string
	}{
		{
			Name:     "manually installed",
			Input:    NewInstallConfig(globalConfig, &tflint.PluginConfig{Name: "foo", Enabled: true}),
			Expected: filepath.Join(globalConfig.PluginDir, "tflint-ruleset-foo"+fileExt()),
		},
		{
			Name: "auto installed",
			Input: NewInstallConfig(globalConfig, &tflint.PluginConfig{
				Name:    "bar",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-bar",
				Version: "0.1.0",
			}),
			Expected: filepath.Join(globalConfig.PluginDir, "github.com/terraform-linters/tflint-ruleset-bar", "0.1.0", "tflint-ruleset-bar"+fileExt()),
		},
	}

	for _, tc := range cases {
		got, err := FindPluginPath(tc.Input)
		if err != nil {
			t.Fatalf("Unexpected error occurred %s", err)
		}
		if got != tc.Expected {
			t.Fatalf("Failed `%s`: want=%s got=%s", tc.Name, tc.Expected, got)
		}
	}
}

func Test_FindPluginPath_withoutExtensionInWindows(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugins")
	defer func() { PluginRoot = original }()

	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{Name: "baz", Enabled: true})
	expected := filepath.Join(PluginRoot, "tflint-ruleset-baz")

	got, err := FindPluginPath(config)
	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}
	if got != expected {
		t.Fatalf("Failed: want=%s got=%s", expected, got)
	}
}
