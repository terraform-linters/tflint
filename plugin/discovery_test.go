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
	defer os.Chdir(cwd)

	err = os.Chdir(filepath.Join(cwd, "test-fixtures", "locals"))
	if err != nil {
		t.Fatal(err)
	}

	plugin, err := Discovery(&tflint.Config{
		Plugins: map[string]*tflint.PluginConfig{
			"foo": {
				Name:    "foo",
				Enabled: true,
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
	defer os.Chdir(cwd)

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
