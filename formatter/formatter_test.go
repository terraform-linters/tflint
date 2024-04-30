package formatter

import (
	"bytes"
	"errors"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

type testRule struct{}

func (r *testRule) Name() string {
	return "test_rule"
}

func (r *testRule) Enabled() bool {
	return true
}

func (r *testRule) Severity() tflint.Severity {
	return sdk.ERROR
}

func (r *testRule) Link() string {
	return "https://github.com"
}

func TestPrintErrorParallel(t *testing.T) {
	// Disable color
	color.NoColor = true

	tests := []struct {
		name   string
		format string
		err    error
		stderr string
	}{
		{
			name:   "default",
			format: "default",
			err:    errors.New("an error occurred\n\nfailed"),
			stderr: `│ an error occurred

failed
`,
		},
		{
			name:   "JSON",
			format: "json",
			err:    errors.New("an error occurred\n\nfailed"),
			stderr: "", // no errors
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			formatter := &Formatter{
				Stdout: stdout,
				Stderr: stderr,
				Format: test.format,
			}

			formatter.PrintErrorParallel(test.err, map[string][]byte{})

			if diff := cmp.Diff(test.stderr, stderr.String()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestPrintParallel(t *testing.T) {
	tests := []struct {
		name   string
		format string
		before func(*Formatter)
		stdout string
		stderr string
		error  bool
	}{
		{
			name:   "default with errors",
			format: "default",
			before: func(f *Formatter) {
				f.PrintErrorParallel(errors.New("an error occurred"), map[string][]byte{})
				f.PrintErrorParallel(errors.New("failed"), map[string][]byte{})
			},
			stdout: "", // no issues
			stderr: "", // no errors
			error:  true,
		},
		{
			name:   "default without errors",
			format: "default",
			before: func(f *Formatter) {},
			stdout: `1 issue(s) found:

Error: test (test_rule)

  on test.tf line 1:
   (source code not available)

Reference: https://github.com

`,
		},
		{
			name:   "JSON with errors",
			format: "json",
			before: func(f *Formatter) {
				f.PrintErrorParallel(errors.New("an error occurred"), map[string][]byte{})
				f.PrintErrorParallel(errors.New("failed"), map[string][]byte{})
			},
			stdout: `{"issues":[{"rule":{"name":"test_rule","severity":"error","link":"https://github.com"},"message":"test","range":{"filename":"test.tf","start":{"line":1,"column":1},"end":{"line":1,"column":4}},"callers":[]}],"errors":[{"message":"an error occurred","severity":"error"},{"message":"failed","severity":"error"}]}`,
			error:  true,
		},
		{
			name:   "JSON without errors",
			format: "json",
			before: func(f *Formatter) {},
			stdout: `{"issues":[{"rule":{"name":"test_rule","severity":"error","link":"https://github.com"},"message":"test","range":{"filename":"test.tf","start":{"line":1,"column":1},"end":{"line":1,"column":4}},"callers":[]}],"errors":[]}`,
		},
	}

	issues := tflint.Issues{
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test.tf",
				Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
				End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formatter := &Formatter{
				Stdout: new(bytes.Buffer),
				Stderr: new(bytes.Buffer),
				Format: test.format,
			}
			test.before(formatter)

			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			formatter.Stdout = stdout
			formatter.Stderr = stderr

			err := formatter.PrintParallel(issues, map[string][]byte{})
			if err != nil && test.error == false {
				t.Errorf("unexpected error: %s", err)
			}
			if err == nil && test.error == true {
				t.Errorf("expected error but got nil")
			}

			if diff := cmp.Diff(test.stdout, stdout.String()); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(test.stderr, stderr.String()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
