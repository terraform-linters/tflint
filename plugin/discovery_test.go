// +build linux darwin

package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_Find(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	originalPluginRoot := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugins")
	defer func() { PluginRoot = originalPluginRoot }()

	if err := buildPlugin(cwd, PluginRoot); err != nil {
		t.Fatal(err)
	}

	plugins, err := Find(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"example": {
				Name:    "example",
				Enabled: false,
			},
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}

	if len(plugins) != 0 {
		t.Fatal("The plugin rule should be disabled")
	}
}

func Test_Find_errors(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	originalPluginRoot := PluginRoot
	PluginRoot = filepath.Join(cwd, "test-fixtures", "plugins")
	defer func() { PluginRoot = originalPluginRoot }()

	if err := buildPlugin(cwd, PluginRoot); err != nil {
		t.Fatal(err)
	}

	_, err = Find(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"not_found": {
				Name:    "not_found",
				Enabled: true,
			},
		},
	})
	if err == nil {
		t.Fatal("An error should occur, but it did not")
	}

	expected := fmt.Sprintf("Plugin `not_found` not found in %s", PluginRoot)
	if err.Error() != expected {
		t.Fatalf("Expected error does not occur: expected=%s got=%s", expected, err.Error())
	}
}

// No successful test case exists under the plugin package due to the "different version" issue.
// Instead, see `plugin_test.go` in the main package.
func Test_OpenPlugin_errors(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Path     string
		Build    bool
		Expected string
	}{
		{
			Name: "invalid binary",
			Path: filepath.Join(cwd, "test-fixtures", "invalid_binary", "tflint-ruleset-example.so"),
			Expected: fmt.Sprintf(
				"Broken plugin `%s` found: The plugin is invalid format",
				filepath.Join(cwd, "test-fixtures", "invalid_binary", "tflint-ruleset-example.so"),
			),
		},
		{
			Name:  "different version",
			Path:  filepath.Join(cwd, "test-fixtures", "plugins", "tflint-ruleset-example.so"),
			Build: true,
			Expected: fmt.Sprintf(
				"Broken plugin `%s` found: The plugin is built with a different version of TFLint. Should be built with v%s",
				filepath.Join(cwd, "test-fixtures", "plugins", "tflint-ruleset-example.so"),
				tflint.Version,
			),
		},
	}

	for _, tc := range cases {
		if tc.Build {
			if err := buildPlugin(cwd, filepath.Dir(tc.Path)); err != nil {
				t.Fatal(err)
			}
		}

		_, err := OpenPlugin(tc.Path)
		if err == nil {
			t.Fatal("An error should occur, but it did not")
		}

		if err.Error() != tc.Expected {
			t.Fatalf("Expected error does not occur: expected=%s got=%s", tc.Expected, err.Error())
		}
	}
}

func buildPlugin(cwd string, dir string) error {
	err := os.Chdir(dir)
	if err != nil {
		return err
	}

	err = exec.Command("go", "build", "--buildmode", "plugin", "-o", "tflint-ruleset-example.so").Run()
	if err != nil {
		return err
	}

	if err = os.Chdir(cwd); err != nil {
		return err
	}
	return nil
}
