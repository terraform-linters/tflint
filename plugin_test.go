// +build linux darwin

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wata727/tflint/plugin"
)

func Test_PluginFind(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	withPluginRoot(t, func(dir string) {
		if err := buildPlugin(cwd, "plugins", dir); err != nil {
			t.Fatal(err)
		}

		plugins, err := plugin.Find()
		if err != nil {
			t.Fatalf("Unexpected error occurred %s", err)
		}

		pluginNames := []string{}
		for _, p := range plugins {
			pluginNames = append(pluginNames, p.Name)
		}

		expected := []string{"plugin_rule"}
		if !cmp.Equal(expected, pluginNames) {
			t.Fatalf("Plugin names are not matched: %s", cmp.Diff(expected, pluginNames))
		}
	})
}

func Test_PluginFind_errors(t *testing.T) {
	cases := []struct {
		Name     string
		Plugin   string
		Expected string
	}{
		{
			Name:     "function not found",
			Plugin:   "invalid_function",
			Expected: "Broken plugin `tflint-ruleset-invalid_function.so` found: The top level `Name` function is undefined",
		},
		{
			Name:     "function signature does not match",
			Plugin:   "invalid_signature",
			Expected: "Broken plugin `tflint-ruleset-invalid_signature.so` found: The top level `Name` function must be of type `func() string`",
		},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		withPluginRoot(t, func(dir string) {
			if err := buildPlugin(cwd, tc.Plugin, dir); err != nil {
				t.Fatal(err)
			}

			_, err := plugin.Find()
			if err == nil {
				t.Fatal("An error should occur, but it did not")
			}

			if err.Error() != tc.Expected {
				t.Fatalf("Expected error does not occur: expected=%s got=%s", tc.Expected, err.Error())
			}
		})
	}
}

func withPluginRoot(t *testing.T, test func(dir string)) {
	dir, err := ioutil.TempDir("", "withPluginRoot")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	originalRoot := plugin.PluginRoot

	plugin.PluginRoot = dir
	test(dir)
	plugin.PluginRoot = originalRoot
}

func buildPlugin(cwd string, name string, to string) error {
	err := os.Chdir(filepath.Join(cwd, "plugin", "test-fixtures", name))
	if err != nil {
		return err
	}

	err = exec.Command("go", "build", "--buildmode", "plugin", "-o", filepath.Join(to, fmt.Sprintf("tflint-ruleset-%s.so", name))).Run()
	if err != nil {
		return err
	}

	if err = os.Chdir(cwd); err != nil {
		return err
	}
	return nil
}
