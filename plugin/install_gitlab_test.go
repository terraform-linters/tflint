package plugin

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
	"golang.org/x/crypto/openpgp"       //nolint:staticcheck
	"golang.org/x/crypto/openpgp/armor" //nolint:staticcheck
)

func TestInstallGitLab(t *testing.T) {
	key, err := openpgp.NewEntity("TFLint test", "", "test@example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	var publicKey bytes.Buffer
	armored, err := armor.Encode(&publicKey, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := key.Serialize(armored); err != nil {
		t.Fatal(err)
	}
	if err := armored.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		signature     string
		omitKey       bool
		missingAsset  string
		corruptZip    bool
		corruptSig    bool
		badChecksums  bool
		badZip        bool
		wrongBinary   bool
		assetStatus   int
		releaseStatus int
		invalidJSON   bool
		duplicate     bool
		wantError     string
		unverified    bool
	}{
		{name: "PGP", signature: "pgp"},
		{name: "auto with signing key"},
		{name: "auto without signing key", omitKey: true, unverified: true},
		{name: "signature disabled", signature: "none", omitKey: true, unverified: true},
		{name: "forced PGP without key", signature: "pgp", omitKey: true, wantError: "No signing key configured"},
		{name: "missing checksums", missingAsset: "checksums.txt", wantError: "Failed to download checksums.txt"},
		{name: "missing signature", missingAsset: "checksums.txt.sig", wantError: "Failed to download checksums.txt.sig"},
		{name: "missing archive", missingAsset: "archive", wantError: "file not found in the GitLab release"},
		{name: "tampered archive", corruptZip: true, wantError: "Failed to verify checksums"},
		{name: "tampered signature", corruptSig: true, wantError: "Failed to check checksums.txt signature"},
		{name: "malformed checksums", badChecksums: true, wantError: "Failed to parse checksums file"},
		{name: "invalid zip", badZip: true, wantError: "Failed to extract binary"},
		{name: "missing binary in zip", wrongBinary: true, wantError: "Does the zip contain"},
		{name: "asset forbidden", assetStatus: http.StatusForbidden, wantError: "release asset download returned 403"},
		{name: "release not found", releaseStatus: http.StatusNotFound, wantError: "GitLab API returned 404"},
		{name: "invalid release JSON", invalidJSON: true, wantError: "Failed to fetch GitLab releases"},
		{name: "duplicate asset", duplicate: true, wantError: "duplicate asset name"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GITLAB_TOKEN", "test-token")
			t.Setenv("GITLAB_TOKEN_gitlab_example_com", "test-token")
			globalConfig := tflint.EmptyConfig()
			globalConfig.PluginDir = t.TempDir()
			config := NewInstallConfig(globalConfig, &tflint.PluginConfig{
				Name: "example", Enabled: true, Version: "0.1.0", SourceType: "gitlab",
				Source:     "gitlab.example.com/group/subgroup/tflint-ruleset-example",
				SourceHost: "gitlab.example.com", SourceOwner: "group/subgroup", SourceRepo: "tflint-ruleset-example",
				Signature: test.signature,
			})
			if !test.omitKey {
				config.SigningKey = publicKey.String()
			}
			binaryName := "tflint-ruleset-example" + fileExt()
			if test.wrongBinary {
				binaryName = "wrong-binary"
			}
			var archive bytes.Buffer
			writer := zip.NewWriter(&archive)
			file, err := writer.Create(binaryName)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(file, "test plugin"); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			archiveBytes := archive.Bytes()
			if test.badZip {
				archiveBytes = []byte("invalid zip")
			}
			checksum := []byte(fmt.Sprintf("%x  %s\n", sha256.Sum256(archiveBytes), config.AssetName()))
			if test.badChecksums {
				checksum = []byte("invalid checksum entry\n")
			}
			var signature bytes.Buffer
			if err := openpgp.DetachSign(&signature, key, bytes.NewReader(checksum), nil); err != nil {
				t.Fatal(err)
			}
			if test.corruptSig {
				signature.Reset()
			}
			if test.corruptZip {
				archiveBytes = []byte("tampered archive")
			}
			files := map[string][]byte{
				"checksums.txt": checksum, "checksums.txt.sig": signature.Bytes(), config.AssetName(): archiveBytes,
			}
			if test.missingAsset == "archive" {
				delete(files, config.AssetName())
			} else {
				delete(files, test.missingAsset)
			}
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("PRIVATE-TOKEN") != "test-token" || r.Header.Get("JOB-TOKEN") != "" {
					t.Error("incorrect GitLab authentication")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if strings.HasPrefix(r.URL.Path, "/api/v4/projects/") {
					if r.URL.EscapedPath() != "/api/v4/projects/group%2Fsubgroup%2Ftflint-ruleset-example/releases/v0.1.0" {
						t.Errorf("unexpected release path: %s", r.URL.EscapedPath())
					}
					if test.releaseStatus != 0 {
						w.WriteHeader(test.releaseStatus)
						return
					}
					if test.invalidJSON {
						_, _ = io.WriteString(w, "invalid JSON")
						return
					}
					links := []map[string]string{}
					for name := range files {
						links = append(links, map[string]string{"name": name, "url": "https://gitlab.example.com/assets/" + name})
					}
					if test.duplicate {
						links = append(links, links[0])
					}
					if err := json.NewEncoder(w).Encode(map[string]any{"assets": map[string]any{"links": links}}); err != nil {
						t.Error(err)
					}
					return
				}
				if test.assetStatus != 0 {
					w.WriteHeader(test.assetStatus)
					return
				}
				_, _ = w.Write(files[strings.TrimPrefix(r.URL.Path, "/assets/")])
			}))
			t.Cleanup(server.Close)
			routeGitLabRequests(t, server)

			path, err := config.Install()
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("got %v, want error containing %q", err, test.wantError)
				}
				if _, err := FindPluginPath(config); !os.IsNotExist(err) {
					t.Fatalf("failed installation must not leave a plugin: %v", err)
				}
				return
			}
			if test.unverified {
				if !errors.Is(err, ErrPluginNotVerified) {
					t.Fatalf("got %v, want ErrPluginNotVerified", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			wantPath := filepath.Join(globalConfig.PluginDir, config.Source, config.Version, "tflint-ruleset-example"+fileExt())
			if path != wantPath {
				t.Fatalf("got install path %q, want %q", path, wantPath)
			}
			found, err := FindPluginPath(config)
			if err != nil || found != path {
				t.Fatalf("installed plugin was not discoverable: %q, %v", found, err)
			}
			contents, err := os.ReadFile(path)
			if err != nil || string(contents) != "test plugin" {
				t.Fatalf("incorrect installed binary: %q, %v", contents, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
				t.Fatal("installed binary is not executable")
			}
		})
	}
}

// Route test hostnames to the TLS server while preserving their original Host header.
func routeGitLabRequests(t *testing.T, server *httptest.Server) {
	t.Helper()
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = gitLabTestTransport(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		return server.Client().Transport.RoundTrip(req)
	})
	t.Cleanup(func() { http.DefaultTransport = original })
}

type gitLabTestTransport func(*http.Request) (*http.Response, error)

func (f gitLabTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGitLabAuthentication(t *testing.T) {
	tests := []struct {
		name       string
		envs       map[string]string
		assetURL   string
		redirect   string
		wantHeader string
		wantToken  string
		wantError  bool
	}{
		{name: "public release"},
		{name: "access token", envs: map[string]string{"GITLAB_TOKEN": "global"}, wantHeader: "PRIVATE-TOKEN", wantToken: "global"},
		{name: "host token takes precedence", envs: map[string]string{"GITLAB_TOKEN": "global", "GITLAB_TOKEN_gitlab_example_com": "host"}, wantHeader: "PRIVATE-TOKEN", wantToken: "host"},
		{name: "empty host token falls back to access token", envs: map[string]string{"GITLAB_TOKEN": "global", "GITLAB_TOKEN_gitlab_example_com": ""}, wantHeader: "PRIVATE-TOKEN", wantToken: "global"},
		{name: "unrelated host token", envs: map[string]string{"GITLAB_TOKEN_other_example_com": "other"}},
		{name: "GitHub token is not used", envs: map[string]string{"GITHUB_TOKEN": "github"}},
		{name: "CI job token", envs: map[string]string{"CI_JOB_TOKEN": "job", "CI_SERVER_URL": "https://gitlab.example.com"}, wantHeader: "JOB-TOKEN", wantToken: "job"},
		{name: "explicit token precedes job token", envs: map[string]string{"GITLAB_TOKEN": "explicit", "CI_JOB_TOKEN": "job", "CI_SERVER_URL": "https://gitlab.example.com"}, wantHeader: "PRIVATE-TOKEN", wantToken: "explicit"},
		{name: "job token for another server", envs: map[string]string{"CI_JOB_TOKEN": "job", "CI_SERVER_URL": "https://other.example.com"}},
		{name: "job token for another port", envs: map[string]string{"CI_JOB_TOKEN": "job", "CI_SERVER_URL": "https://gitlab.example.com:8443"}},
		{name: "job token without server", envs: map[string]string{"CI_JOB_TOKEN": "job"}},
		{name: "direct external asset", envs: map[string]string{"GITLAB_TOKEN": "private"}, assetURL: "https://cdn.example.com/asset"},
		{name: "direct subdomain asset", envs: map[string]string{"GITLAB_TOKEN": "private"}, assetURL: "https://cdn.gitlab.example.com/asset"},
		{name: "direct asset on another port", envs: map[string]string{"GITLAB_TOKEN": "private"}, assetURL: "https://gitlab.example.com:8443/asset"},
		{name: "same origin redirect", envs: map[string]string{"GITLAB_TOKEN": "private"}, redirect: "https://gitlab.example.com/asset", wantHeader: "PRIVATE-TOKEN", wantToken: "private"},
		{name: "cross origin redirect", envs: map[string]string{"GITLAB_TOKEN": "private"}, redirect: "https://cdn.example.com/asset"},
		{name: "subdomain redirect", envs: map[string]string{"GITLAB_TOKEN": "private"}, redirect: "https://cdn.gitlab.example.com/asset"},
		{name: "job token redirect", envs: map[string]string{"CI_JOB_TOKEN": "job", "CI_SERVER_URL": "https://gitlab.example.com"}, redirect: "https://cdn.example.com/asset"},
		{name: "HTTP asset rejected", envs: map[string]string{"GITLAB_TOKEN": "private"}, assetURL: "http://gitlab.example.com/asset", wantError: true},
		{name: "downgrade redirect rejected", envs: map[string]string{"GITLAB_TOKEN": "private"}, redirect: "http://gitlab.example.com/asset", wantError: true},
		{name: "URL credentials rejected", assetURL: "https://user:password@gitlab.example.com/asset", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, variable := range []string{"GITLAB_TOKEN", "GITLAB_TOKEN_gitlab_example_com", "GITLAB_TOKEN_other_example_com", "CI_JOB_TOKEN", "CI_SERVER_URL", "GITHUB_TOKEN"} {
				t.Setenv(variable, "")
			}
			for variable, value := range test.envs {
				t.Setenv(variable, value)
			}
			requests := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.URL.Path == "/redirect" {
					header, token := "PRIVATE-TOKEN", test.envs["GITLAB_TOKEN"]
					if token == "" {
						header, token = "JOB-TOKEN", test.envs["CI_JOB_TOKEN"]
					}
					if r.Header.Get(header) != token {
						t.Error("initial redirect request was not authenticated")
					}
					http.Redirect(w, r, test.redirect, http.StatusFound)
					return
				}
				if test.wantError {
					t.Error("unsafe request reached the server")
				}
				for _, header := range []string{"PRIVATE-TOKEN", "JOB-TOKEN", "Authorization"} {
					want := ""
					if header == test.wantHeader {
						want = test.wantToken
					}
					if r.Header.Get(header) != want {
						t.Errorf("unexpected %s header on %s", header, r.Host)
					}
				}
				_, _ = io.WriteString(w, "asset")
			}))
			t.Cleanup(server.Close)
			routeGitLabRequests(t, server)
			config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{SourceType: "gitlab", SourceHost: "gitlab.example.com"})
			assetURL := test.assetURL
			if assetURL == "" {
				assetURL = "https://gitlab.example.com/asset"
			}
			if test.redirect != "" {
				assetURL = "https://gitlab.example.com/redirect"
			}
			file, err := downloadGitLabAsset(t.Context(), newGitLabClient(config), assetURL)
			if file != nil {
				defer os.Remove(file.Name())
				defer file.Close()
			}
			if test.wantError {
				if err == nil || !strings.Contains(err.Error(), "require HTTPS URLs without user information") {
					t.Fatalf("got %v, want rejected unsafe URL", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantRequests := 1
			if test.redirect != "" {
				wantRequests++
			}
			if requests != wantRequests {
				t.Fatalf("got %d requests, want %d", requests, wantRequests)
			}
		})
	}
}
