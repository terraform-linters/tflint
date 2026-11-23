package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
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
	if plugin != nil {
		defer plugin.Clean()
	}

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

	t.Chdir(filepath.Join(cwd, "test-fixtures", "locals"))

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
	if plugin != nil {
		defer plugin.Clean()
	}

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

	t.Setenv("TFLINT_PLUGIN_DIR", filepath.Join(cwd, "test-fixtures", "locals", ".tflint.d", "plugins"))

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
	if plugin != nil {
		defer plugin.Clean()
	}

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
	if plugin != nil {
		defer plugin.Clean()
	}

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

	t.Chdir(filepath.Join(cwd, "test-fixtures", "no_plugins"))

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
	expected := fmt.Sprintf(`Plugin "foo" not found in %s`, PluginRoot)
	if err.Error() != expected {
		t.Fatalf("The error message is not matched: want=%s, got=%s", expected, err.Error())
	}
}

func Test_Discovery_plugin_name_is_directory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(filepath.Join(cwd, "test-fixtures", "plugin_name_is_directory"))

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
	expected := fmt.Sprintf(`Plugin "foo" not found in %s`, PluginRoot)
	if err.Error() != expected {
		t.Fatalf("The error message is not matched: want=%s, got=%s", expected, err.Error())
	}
}

func Test_Discovery_notFoundForAutoInstallation(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(filepath.Join(cwd, "test-fixtures", "no_plugins"))

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
	expected := `Plugin "foo" not found. Did you run "tflint --init"?`
	if err.Error() != expected {
		t.Fatalf("Error message not matched: want=%s, got=%s", expected, err.Error())
	}
}

func Test_Discovery_bundledPluginWithVersion(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(filepath.Join(cwd, "test-fixtures", "no_plugins"))

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "no_plugins")
	defer func() { PluginRoot = original }()

	_, err = Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"terraform": {
				Name:    "terraform",
				Enabled: true,
				Source:  "github.com/terraform-linters/tflint-ruleset-terraform",
				Version: "0.1.0",
			},
		},
	})

	if err == nil {
		t.Fatal("An error should have occurred, but it did not occur")
	}
	expected := `Plugin "terraform" not found. Did you run "tflint --init"?`
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
		t.Run(tc.Name, func(t *testing.T) {
			got, err := FindPluginPath(tc.Input)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.Expected {
				t.Errorf("want=%s got=%s", tc.Expected, got)
			}
		})
	}
}

func Test_FindPluginPath_locals(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(filepath.Join(cwd, "test-fixtures", "locals"))

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
		t.Run(tc.Name, func(t *testing.T) {
			got, err := FindPluginPath(tc.Input)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.Expected {
				t.Errorf("want=%s got=%s", tc.Expected, got)
			}
		})
	}
}

func Test_FindPluginPath_envVar(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(cwd, "test-fixtures", "locals", ".tflint.d", "plugins")
	t.Setenv("TFLINT_PLUGIN_DIR", dir)

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
		t.Run(tc.Name, func(t *testing.T) {
			got, err := FindPluginPath(tc.Input)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.Expected {
				t.Errorf("want=%s got=%s", tc.Expected, got)
			}
		})
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
		t.Run(tc.Name, func(t *testing.T) {
			got, err := FindPluginPath(tc.Input)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.Expected {
				t.Errorf("want=%s got=%s", tc.Expected, got)
			}
		})
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

func Test_Discovery_lockfile(t *testing.T) {
	tflint.DisableBundledPlugin = true
	defer func() { tflint.DisableBundledPlugin = false }()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	original := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugins")
	defer func() { PluginRoot = original }()

	config := `
plugin "bar" {
  enabled = %t
  source  = "github.com/terraform-linters/tflint-ruleset-bar"
  version = "0.1.0"
}
`
	lockfile := `
lockfile_version = %d

plugin "github.com/terraform-linters/tflint-ruleset-bar" {
  version       = "%s"
  repository_id = 123
  hashes        = {}
}
`

	tests := []struct {
		name     string
		enabled  bool
		lockfile string
		wantErr  string
	}{
		{
			name:    "no lockfile",
			enabled: true,
		},
		{
			name:     "locked",
			enabled:  true,
			lockfile: fmt.Sprintf(lockfile, 1, "0.1.0"),
		},
		{
			name:     "not locked",
			enabled:  true,
			lockfile: "lockfile_version = 1\n",
			wantErr:  `Plugin "bar" (source: github.com/terraform-linters/tflint-ruleset-bar, version: 0.1.0) is not recorded in the lockfile .tflint.lock.hcl. Did you run "tflint --init"?`,
		},
		{
			name:     "not locked in disabled plugin",
			enabled:  false,
			lockfile: "lockfile_version = 1\n",
		},
		{
			name:     "version mismatch",
			enabled:  true,
			lockfile: fmt.Sprintf(lockfile, 1, "0.2.0"),
			wantErr:  `Plugin "bar" version 0.1.0 does not match the lockfile .tflint.lock.hcl (locked: 0.2.0). Did you run "tflint --init"?`,
		},
		{
			name:     "version mismatch in disabled plugin",
			enabled:  false,
			lockfile: fmt.Sprintf(lockfile, 1, "0.2.0"),
		},
		{
			name:     "unsupported lockfile version",
			enabled:  true,
			lockfile: fmt.Sprintf(lockfile, 2, "0.1.0"),
			wantErr:  "Please upgrade TFLint.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.WriteFile(".tflint.hcl", fmt.Appendf(nil, config, test.enabled), 0o644); err != nil {
				t.Fatal(err)
			}
			if test.lockfile != "" {
				if err := os.WriteFile(".tflint.lock.hcl", []byte(test.lockfile), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, "")
			if err != nil {
				t.Fatal(err)
			}

			plugin, err := Discovery(cfg)
			if plugin != nil {
				defer plugin.Clean()
			}

			if test.wantErr != "" {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error to contain %q, but got %q", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
