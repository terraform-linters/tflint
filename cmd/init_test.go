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

func fakePluginBinary(name string) []byte {
	return []byte(fmt.Sprintf("%s plugin binary", name))
}

func fakePluginBinaryHash(name string) string {
	sum := sha256.Sum256(fakePluginBinary(name))
	return "h1:" + hex.EncodeToString(sum[:])
}

// fakeGitHubServer starts a fake GitHub Enterprise Server API serving releases of tflint-ruleset-{name}.
// It replaces http.DefaultTransport to trust the test server certificate.
// Returns the server host and a pointer to the number of plugin downloads.
func fakeGitHubServer(t *testing.T, repositoryID *int64) (string, *int) {
	t.Helper()

	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	zipFile := func(name string) []byte {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		f, err := w.Create(fmt.Sprintf("tflint-ruleset-%s%s", name, ext))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(fakePluginBinary(name)); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}
	checksums := func(name string) []byte {
		zipSum := sha256.Sum256(zipFile(name))
		binarySum := sha256.Sum256(fakePluginBinary(name))
		return fmt.Appendf(nil,
			"%s  tflint-ruleset-%s_%s.zip\n%s  tflint-ruleset-%s_%s\n",
			hex.EncodeToString(zipSum[:]), name, platform,
			hex.EncodeToString(binarySum[:]), name, platform,
		)
	}
	pluginName := func(r *http.Request) string {
		return strings.TrimPrefix(r.PathValue("repo"), "tflint-ruleset-")
	}

	downloads := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": *repositoryID})
	})
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}/releases/tags/{tag}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"assets": []map[string]any{
				{"id": 1, "name": "checksums.txt"},
				{"id": 2, "name": fmt.Sprintf("tflint-ruleset-%s_%s.zip", pluginName(r), platform)},
			},
		})
	})
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}/releases/assets/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(checksums(pluginName(r)))
	})
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}/releases/assets/2", func(w http.ResponseWriter, r *http.Request) {
		downloads++
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(zipFile(pluginName(r)))
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	original := http.DefaultTransport
	http.DefaultTransport = server.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = original })

	return strings.TrimPrefix(server.URL, "https://"), &downloads
}

func TestInitLockfile(t *testing.T) {
	tflint.DisableBundledPlugin = true
	defer func() { tflint.DisableBundledPlugin = false }()

	repositoryID := int64(123)
	host, downloads := fakeGitHubServer(t, &repositoryID)
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	pluginDir := t.TempDir()
	t.Setenv("TFLINT_PLUGIN_DIR", pluginDir)
	installedPath := filepath.Join(pluginDir, host, "owner", "tflint-ruleset-foo", "0.1.0", "tflint-ruleset-foo")
	if runtime.GOOS == "windows" {
		installedPath += ".exe"
	}
	removeInstalledPlugins := func(t *testing.T) {
		t.Helper()
		if err := os.RemoveAll(filepath.Join(pluginDir, host)); err != nil {
			t.Fatal(err)
		}
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
		*downloads = 0
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli, err := NewCLI(outStream, errStream)
		if err != nil {
			t.Fatal(err)
		}
		status := cli.Run(append([]string{"./tflint"}, args...))
		return status, outStream.String(), errStream.String()
	}
	runInit := func(t *testing.T) string {
		t.Helper()
		status, stdout, stderr := run(t, "--init")
		if status != ExitCodeOK {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		return stdout
	}

	pluginConfig := func(name string) string {
		return fmt.Sprintf(`
plugin "%s" {
  enabled   = true
  version   = "0.1.0"
  source    = "%s/owner/tflint-ruleset-%s"
  signature = "none"
}
`, name, host, name)
	}
	lockEntry := func(name string) string {
		return fmt.Sprintf(`
plugin "%s/owner/tflint-ruleset-%s" {
  version       = "0.1.0"
  repository_id = 123
  hashes = {
    %s = "%s"
  }
}
`, host, name, platform, fakePluginBinaryHash(name))
	}
	lockfileHeader := `# This file is maintained automatically by "tflint --init".
# Manual edits may be lost in future updates.

lockfile_version = 1
`

	config := pluginConfig("foo")
	writeFile(t, ".tflint.hcl", config)
	wantLockfile := lockfileHeader + lockEntry("foo")

	t.Run("create lockfile", func(t *testing.T) {
		stdout := runInit(t)
		if !strings.Contains(stdout, "Created lockfile .tflint.lock.hcl") {
			t.Fatalf("expected a lockfile creation log, but got: %s", stdout)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("already installed", func(t *testing.T) {
		stdout := runInit(t)
		if !strings.Contains(stdout, "All plugins are already installed") || strings.Contains(stdout, "lockfile") {
			t.Fatalf("unexpected output: %s", stdout)
		}
		if *downloads != 0 {
			t.Fatalf("expected no downloads, but got %d", *downloads)
		}
	})

	t.Run("reinstall with matched lockfile", func(t *testing.T) {
		removeInstalledPlugins(t)

		stdout := runInit(t)
		if strings.Contains(stdout, "lockfile") {
			t.Fatalf("expected the lockfile not to be updated, but got: %s", stdout)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("reinstall with mismatched hash", func(t *testing.T) {
		removeInstalledPlugins(t)
		tampered := strings.Replace(wantLockfile, fakePluginBinaryHash("foo"), "h1:"+strings.Repeat("0", 64), 1)
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
		if *downloads != 0 {
			t.Fatalf("expected no downloads, but got %d", *downloads)
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		runInit(t)
		writeFile(t, ".tflint.lock.hcl", strings.Replace(wantLockfile, `"0.1.0"`, `"0.0.1"`, 1))

		status, stdout, stderr := run(t)
		if status != ExitCodeError {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stderr, `Plugin "foo" version 0.1.0 does not match the lockfile .tflint.lock.hcl (locked: 0.0.1). Did you run "tflint --init"?`) {
			t.Fatalf("expected a version mismatch error, but got: %s", stderr)
		}

		// The installed plugin is reinstalled to update the lockfile
		stdout = runInit(t)
		if !strings.Contains(stdout, `Reinstalling "foo" plugin to record it in the lockfile...`) || !strings.Contains(stdout, "Updated lockfile .tflint.lock.hcl") {
			t.Fatalf("unexpected output: %s", stdout)
		}
		if *downloads != 1 {
			t.Fatalf("expected 1 download, but got %d", *downloads)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("not recorded in the lockfile", func(t *testing.T) {
		writeFile(t, ".tflint.lock.hcl", lockfileHeader)

		status, stdout, stderr := run(t)
		if status != ExitCodeError {
			t.Fatalf("unexpected status %d: stdout=%s, stderr=%s", status, stdout, stderr)
		}
		if !strings.Contains(stderr, fmt.Sprintf(`Plugin "foo" (source: %s/owner/tflint-ruleset-foo, version: 0.1.0) is not recorded in the lockfile .tflint.lock.hcl. Did you run "tflint --init"?`, host)) {
			t.Fatalf("expected a not recorded error, but got: %s", stderr)
		}

		// The installed plugin is reinstalled to record it in the lockfile
		stdout = runInit(t)
		if !strings.Contains(stdout, `Reinstalling "foo" plugin to record it in the lockfile...`) || !strings.Contains(stdout, "Updated lockfile .tflint.lock.hcl") {
			t.Fatalf("unexpected output: %s", stdout)
		}
		if got := readFile(t, ".tflint.lock.hcl"); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("remove unused entry", func(t *testing.T) {
		writeFile(t, ".tflint.hcl", "")
		defer writeFile(t, ".tflint.hcl", config)

		runInit(t)
		if got := readFile(t, ".tflint.lock.hcl"); got != lockfileHeader {
			t.Fatalf("expected the entry to be removed:\n%s", got)
		}
	})

	t.Run("installed plugin without lockfile", func(t *testing.T) {
		if err := os.Remove(".tflint.lock.hcl"); err != nil {
			t.Fatal(err)
		}

		stdout := runInit(t)
		if !strings.Contains(stdout, "All plugins are already installed") {
			t.Fatalf("unexpected output: %s", stdout)
		}
		if *downloads != 0 {
			t.Fatalf("expected no downloads, but got %d", *downloads)
		}
		if _, err := os.Stat(".tflint.lock.hcl"); !os.IsNotExist(err) {
			t.Fatalf("the lockfile must not be created: %v", err)
		}
	})

	t.Run("add a new plugin without lockfile", func(t *testing.T) {
		writeFile(t, ".tflint.hcl", config+pluginConfig("bar"))
		defer writeFile(t, ".tflint.hcl", config)

		// The installed plugin is also reinstalled to record it in the new lockfile
		stdout := runInit(t)
		for _, want := range []string{
			`Installing "bar" plugin...`,
			`Reinstalling "foo" plugin to record it in the lockfile...`,
			"Created lockfile .tflint.lock.hcl",
		} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("expected the output to contain %q, but got: %s", want, stdout)
			}
		}
		want := lockfileHeader + lockEntry("bar") + lockEntry("foo")
		if got := readFile(t, ".tflint.lock.hcl"); got != want {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})

	t.Run("custom config file", func(t *testing.T) {
		removeInstalledPlugins(t)
		if err := os.Mkdir("configs", 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join("configs", "custom.hcl"), config)
		t.Setenv("TFLINT_CONFIG_FILE", filepath.Join("configs", "custom.hcl"))

		runInit(t)
		if got := readFile(t, filepath.Join("configs", "custom.lock.hcl")); got != wantLockfile {
			t.Fatalf("unexpected lockfile:\n%s", got)
		}
	})
}
