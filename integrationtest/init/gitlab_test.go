package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/cmd"
)

func TestInitGitLab(t *testing.T) {
	t.Chdir(t.TempDir())
	pluginDir := t.TempDir()
	t.Setenv("TFLINT_PLUGIN_DIR", pluginDir)
	t.Setenv("TFLINT_CONFIG_FILE", "")
	if err := os.WriteFile(".tflint.hcl", []byte(`plugin "example" {
  enabled = true
  source_type = "gitlab"
  source = "gitlab.example.com/group/subgroup/ruleset"
  version = "0.1.0"
  signature = "none"
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	binaryName := "tflint-ruleset-example"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	file, err := writer.Create(binaryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, "example plugin"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	archiveName := fmt.Sprintf("tflint-ruleset-example_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
	checksum := fmt.Sprintf("%x  %s\n", sha256.Sum256(archive.Bytes()), archiveName)
	requests := 0
	original := http.DefaultTransport
	http.DefaultTransport = gitLabInitTransport(func(req *http.Request) (*http.Response, error) {
		requests++
		var body []byte
		switch req.URL.EscapedPath() {
		case "/api/v4/projects/group%2Fsubgroup%2Fruleset/releases/v0.1.0":
			body = []byte(fmt.Sprintf(`{"assets":{"links":[{"name":"checksums.txt","url":"https://gitlab.example.com/checksums.txt"},{"name":%q,"url":"https://gitlab.example.com/archive.zip"}]}}`, archiveName))
		case "/checksums.txt":
			body = []byte(checksum)
		case "/archive.zip":
			body = archive.Bytes()
		default:
			t.Fatalf("unexpected request: %s", req.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header), Request: req}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = original })
	var out, stderr bytes.Buffer
	cli, err := cmd.NewCLI(&out, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if code := cli.Run([]string{"--init"}); code != cmd.ExitCodeOK {
		t.Fatalf("init exited %d: %s", code, &stderr)
	}
	if !strings.Contains(out.String(), `Installed "example" (source: gitlab.example.com/group/subgroup/ruleset, version: 0.1.0)`) {
		t.Fatalf("missing installation output: %s", &out)
	}
	path := filepath.Join(pluginDir, "gitlab.example.com/group/subgroup/ruleset/0.1.0", binaryName)
	contents, err := os.ReadFile(path)
	if err != nil || string(contents) != "example plugin" {
		t.Fatalf("incorrect installed plugin: %q, %v", contents, err)
	}
	if requests != 3 {
		t.Fatalf("got %d requests, want release metadata, checksums and archive", requests)
	}
	out.Reset()
	if code := cli.Run([]string{"--init"}); code != cmd.ExitCodeOK {
		t.Fatalf("second init exited %d: %s", code, &stderr)
	}
	if !strings.Contains(out.String(), "All plugins are already installed") || requests != 3 {
		t.Fatalf("second init must reuse the installed plugin: requests=%d, output=%s", requests, &out)
	}
}

type gitLabInitTransport func(*http.Request) (*http.Response, error)

func (f gitLabInitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
