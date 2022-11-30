package tflint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform"
)

// TestRunner returns a runner for testing.
// Note that this runner ignores a config, annotations, and input variables.
func TestRunner(t *testing.T, files map[string]string) *Runner {
	return TestRunnerWithConfig(t, files, EmptyConfig())
}

// TestRunnerWithConfig returns a runner with passed config for testing.
func TestRunnerWithConfig(t *testing.T, files map[string]string, config *Config) *Runner {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	for name, src := range files {
		err := fs.WriteFile(name, []byte(src), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	loader, err := terraform.NewLoader(fs)
	if err != nil {
		t.Fatal(err)
	}

	dirMap := map[string]*struct{}{}
	for file := range files {
		dirMap[filepath.Dir(file)] = nil
	}
	dirs := make([]string, 0)
	for dir := range dirMap {
		dirs = append(dirs, dir)
	}

	if len(dirs) > 1 {
		t.Fatalf("All test files must be in the same directory, got %d directories: %v", len(dirs), dirs)
		return nil
	}

	var dir string
	if len(dirs) == 0 {
		dir = "."
	} else {
		dir = dirs[0]
	}

	configs, diags := loader.LoadConfig(dir, config.Module)
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	runner, err := NewRunner(config, map[string]Annotations{}, configs, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}

	return runner
}
