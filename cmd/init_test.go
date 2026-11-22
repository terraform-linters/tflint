package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

// fakeGitHubServer starts a fake GitHub Enterprise Server API serving releases of tflint-ruleset-foo.
// It replaces http.DefaultTransport to trust the test server certificate.
func fakeGitHubServer(t *testing.T, repositoryID *int64, binary []byte) string {
	t.Helper()

	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	binaryName := "tflint-ruleset-foo"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	zipBuf := new(bytes.Buffer)
	w := zip.NewWriter(zipBuf)
	f, err := w.Create(binaryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(binary); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	zipSum := sha256.Sum256(zipBuf.Bytes())
	binarySum := sha256.Sum256(binary)
	checksums := fmt.Sprintf(
		"%s  tflint-ruleset-foo_%s.zip\n%s  tflint-ruleset-foo_%s\n",
		hex.EncodeToString(zipSum[:]), platform,
		hex.EncodeToString(binarySum[:]), platform,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v3/repos/owner/tflint-ruleset-foo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": *repositoryID})
	})
	mux.HandleFunc("GET /api/v3/repos/owner/tflint-ruleset-foo/releases/tags/{tag}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"assets": []map[string]any{
				{"id": 1, "name": "checksums.txt"},
				{"id": 2, "name": fmt.Sprintf("tflint-ruleset-foo_%s.zip", platform)},
			},
		})
	})
	mux.HandleFunc("GET /api/v3/repos/owner/tflint-ruleset-foo/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte(checksums))
	})
	mux.HandleFunc("GET /api/v3/repos/owner/tflint-ruleset-foo/releases/assets/2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipBuf.Bytes())
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	original := http.DefaultTransport
	http.DefaultTransport = server.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = original })

	return strings.TrimPrefix(server.URL, "https://")
}

func TestInitLockfile(t *testing.T) {
	tflint.DisableBundledPlugin = true
	defer func() { tflint.DisableBundledPlugin = false }()

	repositoryID := int64(123)
	binary := []byte("plugin binary")
	host := fakeGitHubServer(t, &repositoryID, binary)
	source := fmt.Sprintf("%s/owner/tflint-ruleset-foo", host)

	sum := sha256.Sum256(binary)
	binaryHash := "h1:" + hex.EncodeToString(sum[:])
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	pluginDir := t.TempDir()
	t.Setenv("TFLINT_PLUGIN_DIR", pluginDir)
	installedPath := filepath.Join(pluginDir, source, "0.1.0", "tflint-ruleset-foo")
	if runtime.GOOS == "windows" {
		installedPath += ".exe"
	}

	t.Chdir(t.TempDir())

	writeFile := func(t *testing.T, name string, content string) {
		t.Helper()
		if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	readFile := func(t *testing.T, name string) string {
		t.Helper()
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	run := func(t *testing.T, args ...string) (int, string, string) {
		t.Helper()
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli, err := NewCLI(outStream, errStream)
		if err != nil {
			t.Fatal(err)
		}
		status := cli.Run(append([]string{"./tflint"}, args...))
		return status, outStream.String(), errStream.String()
	}

	config := fmt.Sprintf(`
plugin "foo" {
  enabled   = true
  version   = "0.1.0"
  source    = "%s"
  signature = "none"
}
`, source)
	writeFile(t, ".tflint.hcl", config)

	wantLockfile := fmt.Sprintf(`# This file is maintained automatically by "tflint --init".
# Manual edits may be lost in future updates.

lockfile_version = 1

plugin "%s" {
  version       = "0.1.0"
  repository_id = 123
  hashes = {
    %s = "%s"
  }
}
`, source, platform, binaryHash)

	t.Run("create lockfile", func(t *testing.T) {
		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stdout, "Created lockfile .tflint.lock.hcl") {
			t.Fatalf("expected a lockfile creation log, but got: %s", stdout)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("already installed", func(t *testing.T) {
		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stdout, "All plugins are already installed") || strings.Contains(stdout, "lockfile") {
			t.Fatalf("unexpected output: %s", stdout)
		}
	})

	t.Run("reinstall with matched lockfile", func(t *testing.T) {
		if err := os.RemoveAll(filepath.Join(pluginDir, host)); err != nil {
			t.Fatal(err)
		}

		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if strings.Contains(stdout, "lockfile") {
			t.Fatalf("expected the lockfile not to be updated, but got: %s", stdout)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("reinstall with mismatched hash", func(t *testing.T) {
		if err := os.RemoveAll(filepath.Join(pluginDir, host)); err != nil {
			t.Fatal(err)
		}
		tampered := strings.Replace(wantLockfile, binaryHash, "h1:"+strings.Repeat("0", 64), 1)
		writeFile(t, ".tflint.lock.hcl", tampered)
		defer writeFile(t, ".tflint.lock.hcl", wantLockfile)

		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeError {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stderr, "does not match the lockfile") {
			t.Fatalf("expected a hash mismatch error, but got: %s", stderr)
		}
		if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
			t.Fatalf("the plugin must not be installed: %v", err)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != tampered {
			t.Fatalf("the lockfile must not be changed:\n%s", got)
		}
	})

	t.Run("reinstall with mismatched repository ID", func(t *testing.T) {
		repositoryID = 456
		defer func() { repositoryID = 123 }()

		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeError {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stderr, "does not match the lockfile: locked=123, actual=456") {
			t.Fatalf("expected a repository ID mismatch error, but got: %s", stderr)
		}
		if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
			t.Fatalf("the plugin must not be installed: %v", err)
		}
	})

	t.Run("version mismatch on inspection", func(t *testing.T) {
		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		writeFile(t, ".tflint.lock.hcl", strings.Replace(wantLockfile, `"0.1.0"`, `"0.0.1"`, 1))

		status, stdout, stderr = run(t)
		if status != ExitCodeError {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stderr, `Plugin "foo" version 0.1.0 does not match the lockfile .tflint.lock.hcl (locked: 0.0.1). Did you run "tflint --init"?`) {
			t.Fatalf("expected a version mismatch error, but got: %s", stderr)
		}
	})

	t.Run("remove stale entry of installed plugin", func(t *testing.T) {
		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stdout, "Updated lockfile .tflint.lock.hcl") {
			t.Fatalf("expected a lockfile update log, but got: %s", stdout)
		}
		want := `# This file is maintained automatically by "tflint --init".
# Manual edits may be lost in future updates.

lockfile_version = 1
`
		if got := readFile(t, ".tflint.lock.hcl"); got != want {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("remove unused entry", func(t *testing.T) {
		writeFile(t, ".tflint.lock.hcl", wantLockfile)
		writeFile(t, ".tflint.hcl", "")
		defer writeFile(t, ".tflint.hcl", config)

		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if strings.Contains(readFile(t, ".tflint.lock.hcl"), "plugin") {
			t.Fatalf("expected the entry to be removed:\n%s", readFile(t, ".tflint.lock.hcl"))
		}
	})

	t.Run("installed plugin without lockfile", func(t *testing.T) {
		if err := os.Remove(".tflint.lock.hcl"); err != nil {
			t.Fatal(err)
		}

		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if _, err := os.Stat(".tflint.lock.hcl"); !os.IsNotExist(err) {
			t.Fatalf("the lockfile must not be created: %v", err)
		}
	})

	t.Run("custom config file", func(t *testing.T) {
		if err := os.RemoveAll(filepath.Join(pluginDir, host)); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir("configs", 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join("configs", "custom.hcl"), config)
		t.Setenv("TFLINT_CONFIG_FILE", filepath.Join("configs", "custom.hcl"))

		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if got := readFile(t, filepath.Join("configs", "custom.lock.hcl")); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
		if _, err := os.Stat(".tflint.lock.hcl"); !os.IsNotExist(err) {
			t.Fatalf("the default lockfile must not be created: %v", err)
		}
	})
}
