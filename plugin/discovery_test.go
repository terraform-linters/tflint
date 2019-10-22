// +build linux darwin

package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_Find(t *testing.T) {
	cases := []struct {
		Name       string
		PluginRoot string
		Expected   []string
	}{
		{
			Name:       "no plugin dir",
			PluginRoot: "not_found",
			Expected:   []string{},
		},
		{
			Name:       "no plugins",
			PluginRoot: "no_plugins",
			Expected:   []string{},
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		originalPluginRoot := PluginRoot
		PluginRoot = filepath.Join(cwd, "test-fixtures", tc.PluginRoot)

		plugins, err := Find()
		if err != nil {
			t.Fatalf("Failed %s: unexpected error occurred %s", tc.Name, err)
		}

		if len(plugins) != 0 {
			t.Fatalf("Failed %s: The plugin rule should not be obtained", tc.Name)
		}
		PluginRoot = originalPluginRoot
	}
}

func Test_Find_errors(t *testing.T) {
	cases := []struct {
		Name       string
		PluginRoot string
		Build      bool
		Expected   string
	}{
		{
			Name:       "invalid binary",
			PluginRoot: "invalid_binary",
			Expected:   "Broken plugin `tflint-ruleset-example.so` found: The plugin is invalid format",
		},
		{
			Name:       "different version",
			PluginRoot: "plugins",
			Build:      true,
			Expected:   fmt.Sprintf("Broken plugin `tflint-ruleset-example.so` found: The plugin is built with a different version of TFLint. Should be built with v%s", tflint.Version),
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		originalPluginRoot := PluginRoot
		PluginRoot = filepath.Join(cwd, "test-fixtures", tc.PluginRoot)

		if tc.Build {
			if err := buildPlugin(cwd, PluginRoot); err != nil {
				t.Fatal(err)
			}
		}

		_, err := Find()
		if err == nil {
			t.Fatal("An error should occur, but it did not")
		}

		if err.Error() != tc.Expected {
			t.Fatalf("Expected error does not occur: expected=%s got=%s", tc.Expected, err.Error())
		}
		PluginRoot = originalPluginRoot
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
