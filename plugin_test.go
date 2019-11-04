// +build linux darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wata727/tflint/plugin"
)

func Test_PluginOpen(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(cwd, "plugin", "test-fixtures", "plugins")

	if err := buildPlugin(cwd, dir); err != nil {
		t.Fatal(err)
	}

	plugin, err := plugin.OpenPlugin(filepath.Join(dir, "tflint-ruleset-example.so"))
	if err != nil {
		t.Fatalf("Unexpected error occurred %s", err)
	}

	if plugin == nil {
		t.Fatal("Cannot got plugin")
	}

	expected := "plugin_rule"
	if plugin.Name != expected {
		t.Fatalf("Plugin rule should be %s, but got %s", expected, plugin.Name)
	}
}

func Test_PluginOpen_errors(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Path     string
		Expected string
	}{
		{
			Name: "function not found",
			Path: filepath.Join(cwd, "plugin", "test-fixtures", "invalid_function", "tflint-ruleset-example.so"),
			Expected: fmt.Sprintf(
				"Broken plugin `%s` found: The top level `Name` function is undefined",
				filepath.Join(cwd, "plugin", "test-fixtures", "invalid_function", "tflint-ruleset-example.so"),
			),
		},
		{
			Name: "function signature does not match",
			Path: filepath.Join(cwd, "plugin", "test-fixtures", "invalid_signature", "tflint-ruleset-example.so"),
			Expected: fmt.Sprintf(
				"Broken plugin `%s` found: The top level `Name` function must be of type `func() string`",
				filepath.Join(cwd, "plugin", "test-fixtures", "invalid_signature", "tflint-ruleset-example.so"),
			),
		},
	}

	for _, tc := range cases {
		if err := buildPlugin(cwd, filepath.Dir(tc.Path)); err != nil {
			t.Fatal(err)
		}

		_, err := plugin.OpenPlugin(tc.Path)
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
