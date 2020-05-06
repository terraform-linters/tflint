package tflint

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
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

	cfg, err := loader.LoadConfig(".")
	if err != nil {
		t.Fatal(err)
	}

	runner, err := NewRunner(config, map[string]Annotations{}, cfg, map[string]*terraform.InputValue{})
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

// AssertAppError is an assertion helper for comparing tflint.Error
func AssertAppError(t *testing.T, expected Error, got error) {
	if appErr, ok := got.(*Error); ok {
		if appErr == nil {
			t.Fatalf("expected err is `%s`, but nothing occurred", expected.Error())
		}
		if appErr.Code != expected.Code {
			t.Fatalf("expected error code is `%s`, but get `%s`", expected.Code, appErr.Code)
		}
		if appErr.Level != expected.Level {
			t.Fatalf("expected error level is `%s`, but get `%s`", expected.Level, appErr.Level)
		}
		if appErr.Error() != expected.Error() {
			t.Fatalf("expected error is `%s`, but get `%s`", expected.Error(), appErr.Error())
		}
	} else {
		t.Fatalf("unexpected error occurred: %s", got)
	}
}

// ruleComparer returns a Comparer func that checks that two rule interfaces
// have the same underlying type. It does not compare struct fields.
func ruleComparer() cmp.Option {
	return cmp.Comparer(func(x, y Rule) bool {
		return reflect.TypeOf(x) == reflect.TypeOf(y)
	})
}
