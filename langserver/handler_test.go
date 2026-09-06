package langserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/logutils"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
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
