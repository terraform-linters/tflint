package tflint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform/terraform"
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

	loader, err := NewLoader(fs, config)
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

	cfg, err := loader.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	f, err := loader.Files()
	if err != nil {
		t.Fatal(err)
	}

	runner, err := NewRunner(config, f, map[string]Annotations{}, cfg, map[string]*terraform.InputValue{})
	if err != nil {
		t.Fatal(err)
	}

	return runner
}

// AssertIssues is an assertion helper for comparing issues
func AssertIssues(t *testing.T, expected Issues, actual Issues) {
	opts := []cmp.Option{
		// Byte field will be ignored because it's not important in tests such as positions
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		ruleComparer(),
	}
	if !cmp.Equal(expected, actual, opts...) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, actual, opts...))
	}
}

// AssertIssuesWithoutRange is an assertion helper for comparing issues
func AssertIssuesWithoutRange(t *testing.T, expected Issues, actual Issues) {
	opts := []cmp.Option{
		cmpopts.IgnoreFields(Issue{}, "Range"),
		ruleComparer(),
	}
	if !cmp.Equal(expected, actual, opts...) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, actual, opts...))
	}
}

// ruleComparer returns a Comparer func that checks that two rule interfaces
// have the same underlying type. It does not compare struct fields.
func ruleComparer() cmp.Option {
	return cmp.Comparer(func(x, y Rule) bool {
		return reflect.TypeOf(x) == reflect.TypeOf(y)
	})
}
