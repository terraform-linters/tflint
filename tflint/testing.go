package tflint

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/terraform"
	"github.com/spf13/afero"
)

// TestRunner returns a runner for testing.
// Note that this runner ignores a config, annotations, and input variables.
func TestRunner(t *testing.T, files map[string]string) *Runner {
	config := EmptyConfig()
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
		cmpopts.IgnoreFields(Issue{}, "Rule"),
	}
	if !cmp.Equal(expected, actual, opts...) {
		t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(expected, actual, opts...))
	}
}

// AssertIssuesWithoutRange is an assertion helper for comparing issues
func AssertIssuesWithoutRange(t *testing.T, expected Issues, actual Issues) {
	opts := []cmp.Option{
		cmpopts.IgnoreFields(Issue{}, "Range"),
		cmpopts.IgnoreFields(Issue{}, "Rule"),
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
			t.Fatalf("expected error code is `%d`, but get `%d`", expected.Code, appErr.Code)
		}
		if appErr.Level != expected.Level {
			t.Fatalf("expected error level is `%d`, but get `%d`", expected.Level, appErr.Level)
		}
		if appErr.Error() != expected.Error() {
			t.Fatalf("expected error is `%s`, but get `%s`", expected.Error(), appErr.Error())
		}
	} else {
		t.Fatalf("unexpected error occurred: %s", got)
	}
}
