package plugin

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/tflint"
)

// fakeRelease is a fake GitHub release served by fakeGitHubServer.
type fakeRelease struct {
	repositoryID int64
	pluginName   string
	version      string
	binary       []byte
	// If true, checksums.txt includes binary hashes
	withBinaryHashes bool
	// Override binary hashes in checksums.txt (platform -> sha256 hex)
	binaryHashes map[string]string
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (r *fakeRelease) zip(t *testing.T) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(fmt.Sprintf("tflint-ruleset-%s%s", r.pluginName, fileExt()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(r.binary); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func (r *fakeRelease) checksums(zipFile []byte) []byte {
	lines := []string{
		fmt.Sprintf("%s  tflint-ruleset-%s_%s.zip", sha256Hex(zipFile), r.pluginName, currentPlatform()),
	}
	if r.withBinaryHashes {
		hashes := map[string]string{
			currentPlatform(): sha256Hex(r.binary),
			"other_arch":      sha256Hex([]byte("other platform binary")),
		}
		for platform, hash := range r.binaryHashes {
			hashes[platform] = hash
		}
		for platform, hash := range hashes {
			lines = append(lines, fmt.Sprintf("%s  tflint-ruleset-%s_%s", hash, r.pluginName, platform))
		}
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

// fakeGitHubServer starts a fake GitHub Enterprise Server API serving the given release.
// It replaces http.DefaultTransport to trust the test server certificate.
// Returns the server host and a pointer to the number of asset downloads.
func fakeGitHubServer(t *testing.T, release *fakeRelease) (string, *int) {
	t.Helper()

	zipFile := release.zip(t)
	assets := map[int][]byte{
		1: release.checksums(zipFile),
		2: zipFile,
	}
	downloads := 0

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": release.repositoryID, "private": false})
	})
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}/releases/tags/{tag}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("tag") != "v"+release.version {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"assets": []map[string]any{
				{"id": 1, "name": "checksums.txt"},
				{"id": 2, "name": fmt.Sprintf("tflint-ruleset-%s_%s.zip", release.pluginName, currentPlatform())},
			},
		})
	})
	mux.HandleFunc("GET /api/v3/repos/owner/{repo}/releases/assets/{id}", func(w http.ResponseWriter, r *http.Request) {
		var id int
		if _, err := fmt.Sscanf(r.PathValue("id"), "%d", &id); err != nil {
			http.NotFound(w, r)
			return
		}
		downloads++
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(assets[id])
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	// Route all requests to the test server. The host name is used instead of
	// "127.0.0.1:port" because the source host is part of the plugin install path,
	// and ":" cannot be used in paths on Windows. The test server certificate is
	// valid for example.com.
	transport := server.Client().Transport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	original := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = original })

	return "example.com", &downloads
}

func newFakeInstallConfig(t *testing.T, host string, name string, version string) *InstallConfig {
	t.Helper()
	original := PluginRoot
	PluginRoot = t.TempDir()
	t.Cleanup(func() { PluginRoot = original })
	t.Setenv("TFLINT_PLUGIN_DIR", "")

	return NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        name,
		Enabled:     true,
		Version:     version,
		Source:      fmt.Sprintf("%s/owner/tflint-ruleset-%s", host, name),
		SourceHost:  host,
		SourceOwner: "owner",
		SourceRepo:  "tflint-ruleset-" + name,
		Signature:   string(SignatureModeNone),
	})
}

func TestInstall_lockfile(t *testing.T) {
	binary := []byte("plugin binary")
	binaryHash := "h1:" + sha256Hex(binary)
	otherHash := "h1:" + sha256Hex([]byte("other platform binary"))
	tampered := "h1:" + sha256Hex([]byte("tampered"))

	tests := []struct {
		name    string
		release *fakeRelease
		version string
		// locked is a function to build the lock entry because the source depends on the server host
		locked    func(source string) *LockedPlugin
		want      func(source string) *LockedPlugin
		wantErr   string
		downloads int
	}{
		{
			name:    "no lock entry with binary hashes",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary, withBinaryHashes: true},
			version: "0.1.0",
			want: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
					"other_arch":      otherHash,
				}}
			},
			downloads: 2,
		},
		{
			name:    "no lock entry without binary hashes",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary},
			version: "0.1.0",
			want: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
				}}
			},
			downloads: 2,
		},
		{
			name:    "lock entry matches",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary, withBinaryHashes: true},
			version: "0.1.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
					"other_arch":      otherHash,
				}}
			},
			want: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
					"other_arch":      otherHash,
				}}
			},
			downloads: 2,
		},
		{
			name:    "lock entry without the current platform",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary},
			version: "0.1.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					"other_arch": otherHash,
				}}
			},
			want: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
					"other_arch":      otherHash,
				}}
			},
			downloads: 2,
		},
		{
			name:    "hash for the current platform does not match",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary},
			version: "0.1.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): tampered,
				}}
			},
			wantErr:   fmt.Sprintf("plugin binary for %s does not match the lockfile", currentPlatform()),
			downloads: 2,
		},
		{
			name:    "hash for another platform does not match",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary, withBinaryHashes: true},
			version: "0.1.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
					"other_arch":      tampered,
				}}
			},
			wantErr:   "plugin binary for other_arch does not match the lockfile",
			downloads: 2,
		},
		{
			name:    "repository ID does not match",
			release: &fakeRelease{repositoryID: 456, version: "0.1.0", binary: binary, withBinaryHashes: true},
			version: "0.1.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
				}}
			},
			wantErr:   "does not match the lockfile: locked=123, actual=456",
			downloads: 0,
		},
		{
			name:    "version upgrade",
			release: &fakeRelease{repositoryID: 123, version: "0.2.0", binary: binary},
			version: "0.2.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): tampered,
					"other_arch":      tampered,
				}}
			},
			want: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.2.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
				}}
			},
			downloads: 2,
		},
		{
			name:    "version upgrade with another repository ID",
			release: &fakeRelease{repositoryID: 456, version: "0.2.0", binary: binary},
			version: "0.2.0",
			locked: func(source string) *LockedPlugin {
				return &LockedPlugin{Source: source, Version: "0.1.0", RepositoryID: 123, Hashes: map[string]string{
					currentPlatform(): binaryHash,
				}}
			},
			wantErr:   "does not match the lockfile: locked=123, actual=456",
			downloads: 0,
		},
		{
			name: "binary does not match checksums.txt",
			release: &fakeRelease{repositoryID: 123, version: "0.1.0", binary: binary, withBinaryHashes: true, binaryHashes: map[string]string{
				currentPlatform(): sha256Hex([]byte("tampered")),
			}},
			version:   "0.1.0",
			wantErr:   "Failed to match the plugin binary hash",
			downloads: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.release.pluginName = "foo"
			host, downloads := fakeGitHubServer(t, test.release)
			config := newFakeInstallConfig(t, host, "foo", test.version)

			var locked *LockedPlugin
			if test.locked != nil {
				locked = test.locked(config.Source)
			}

			path, entry, err := config.Install(locked)
			if err == ErrPluginNotVerified {
				err = nil
			}

			if *downloads != test.downloads {
				t.Errorf("expected %d downloads, but got %d", test.downloads, *downloads)
			}

			dir, dirErr := getPluginDir(config.globalConfig)
			if dirErr != nil {
				t.Fatal(dirErr)
			}
			installPath := filepath.Join(dir, config.InstallPath()+fileExt())

			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error to contain %q, but got %q", test.wantErr, err)
				}
				if _, err := os.Stat(installPath); !os.IsNotExist(err) {
					t.Fatalf("the plugin must not be installed, but got %v", err)
				}
				// No temporary files should be left
				entries, _ := os.ReadDir(filepath.Dir(installPath))
				if len(entries) != 0 {
					t.Fatalf("expected no files in the plugin dir, but got %d", len(entries))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if path != installPath {
				t.Errorf("expected path %s, but got %s", installPath, path)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, binary) {
				t.Errorf("installed binary is invalid: %q", got)
			}
			if diff := cmp.Diff(test.want(config.Source), entry); diff != "" {
				t.Errorf("lock entry mismatch: %s", diff)
			}
		})
	}
}
