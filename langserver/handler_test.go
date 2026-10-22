package langserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/logutils"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_uriToPath_windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	uri := lsp.DocumentURI("file:///c%3A/example%20directory")
	value, _ := uriToPath(uri)
	expected := "c:/example directory"

	if !cmp.Equal(expected, value) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, value))
	}
}

func Test_uriToPath_others(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	uri := lsp.DocumentURI("file:///example%20directory")
	value, _ := uriToPath(uri)
	expected := "/example directory"

	if !cmp.Equal(expected, value) {
		t.Fatalf("Diff: %s", cmp.Diff(expected, value))
	}
}

func Test_handle_logsParamsOnlyAtTraceLevel(t *testing.T) {
	secret := "example-sensitive-value"
	source := fmt.Sprintf("locals {\n  api_token = %q\n}\n", secret)
	params := json.RawMessage(fmt.Sprintf(`{"contentChanges":[{"text":%q}]}`, source))
	req := &jsonrpc2.Request{Method: "textDocument/didChange", Params: &params}

	handler := &handler{shutdown: true}

	tests := []struct {
		name       string
		level      string
		wantParams bool
	}{
		{
			name:       "unset",
			level:      "",
			wantParams: false,
		},
		{
			name:       "debug",
			level:      "debug",
			wantParams: false,
		},
		{
			name:       "trace",
			level:      "trace",
			wantParams: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("TFLINT_LOG", test.level)

			logs := captureLogs(t, func() {
				_, _ = handler.handle(context.Background(), nil, req)
			})

			if !strings.Contains(logs, "Received textDocument/didChange") {
				t.Errorf("Expected the method to be logged, but got: %s", logs)
			}
			if logged := strings.Contains(logs, secret); logged != test.wantParams {
				t.Errorf("Expected params logged to be %t, but got %t: %s", test.wantParams, logged, logs)
			}
		})
	}
}

func captureLogs(t *testing.T, f func()) string {
	t.Helper()

	buf := new(bytes.Buffer)
	writer := log.Writer()
	t.Cleanup(func() { log.SetOutput(writer) })
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(os.Getenv("TFLINT_LOG"))),
		Writer:   buf,
	})

	f()

	return buf.String()
}

func Test_hclDiagnosticsToLSP(t *testing.T) {
	root := t.TempDir()
	h := &handler{rootDir: root, diagsPaths: []string{filepath.Join(root, "stale.tf")}}

	diags := hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing expression",
			Detail:   "Expected the start of an expression.",
			Subject: &hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 2, Column: 12},
				End:      hcl.Pos{Line: 3, Column: 1},
			},
		},
	}

	got := h.hclDiagnosticsToLSP(diags)
	path := filepath.Join(root, "main.tf")
	stale := filepath.Join(root, "stale.tf")

	if _, ok := got[stale]; !ok {
		t.Fatalf("expected stale path %q to be cleared", stale)
	}
	if len(got[stale]) != 0 {
		t.Fatalf("expected empty diagnostics for stale path, got %#v", got[stale])
	}
	if len(got[path]) != 1 {
		t.Fatalf("expected 1 diagnostic for %q, got %#v", path, got[path])
	}
	diag := got[path][0]
	if diag.Severity != lsp.Error {
		t.Errorf("severity: got %v want %v", diag.Severity, lsp.Error)
	}
	if diag.Message != "Missing expression; Expected the start of an expression." {
		t.Errorf("message: %q", diag.Message)
	}
	wantRange := lsp.Range{
		Start: lsp.Position{Line: 1, Character: 11},
		End:   lsp.Position{Line: 2, Character: 0},
	}
	if diff := cmp.Diff(wantRange, diag.Range); diff != "" {
		t.Errorf("range mismatch (-want +got):\n%s", diff)
	}
	if len(h.diagsPaths) != 1 || h.diagsPaths[0] != path {
		t.Errorf("diagsPaths not updated: %#v", h.diagsPaths)
	}
}

func Test_inspect_publishesHCLDiagnosticsInsteadOfError(t *testing.T) {
	// Prefer MkdirTemp over t.TempDir: BuildRunners/loader can leave Windows
	// file handles open, and t.TempDir failing RemoveAll fails the test even
	// when assertions passed.
	root, err := os.MkdirTemp("", "tflint-inspect-*")
	if err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
		_ = os.RemoveAll(root)
	})

	// Match textDocument/didChange: overlay writes the basename into the memfs layer.
	h := &handler{
		config:     tflint.EmptyConfig(),
		fs:         afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs()),
		rootDir:    root,
		plugin:     nil,
		diagsPaths: []string{},
	}
	if err := afero.WriteFile(h.fs, "main.tf", []byte("locals {\n  example =\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	diags, err := h.inspect()
	if err != nil {
		t.Fatalf("inspect should not fail the handler on transient HCL errors, got: %v", err)
	}

	found := false
	for path, ds := range diags {
		if len(ds) == 0 {
			continue
		}
		found = true
		if !strings.Contains(ds[0].Message, "Missing expression") && !strings.Contains(ds[0].Message, "Invalid") && !strings.Contains(ds[0].Message, "Argument") {
			// Accept any HCL parse diagnostic for the incomplete locals block.
			t.Logf("diagnostic on %q: %s", path, ds[0].Message)
		}
	}
	if !found {
		t.Fatalf("expected HCL diagnostics to be published, got %#v", diags)
	}
}

func Test_errorsAsType_hclDiagnosticsWrapped(t *testing.T) {
	wrapped := fmt.Errorf("Failed to load the root module; %w", hcl.Diagnostics{
		&hcl.Diagnostic{Severity: hcl.DiagError, Summary: "Missing expression"},
	})
	diags, ok := errors.AsType[hcl.Diagnostics](wrapped)
	if !ok {
		t.Fatal("expected errors.AsType to unwrap hcl.Diagnostics")
	}
	if len(diags) != 1 || diags[0].Summary != "Missing expression" {
		t.Fatalf("unexpected diags: %#v", diags)
	}
}