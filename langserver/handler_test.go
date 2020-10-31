package langserver

import (
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	lsp "github.com/sourcegraph/go-lsp"
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
