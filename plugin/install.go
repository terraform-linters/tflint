package plugin

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/terraform-linters/tflint/tflint"
	"golang.org/x/net/idna"
	"golang.org/x/oauth2"
)

const defaultSourceHost = "github.com"

// InstallConfig is a config for plugin installation.
// This is a wrapper for PluginConfig and manages naming conventions
// and directory names for installation.
// Note that need a global config to manage installation directory.
type InstallConfig struct {
	globalConfig *tflint.Config

	*tflint.PluginConfig
}

// NewInstallConfig returns a new InstallConfig from passed PluginConfig.
func NewInstallConfig(config *tflint.Config, pluginCfg *tflint.PluginConfig) *InstallConfig {
	return &InstallConfig{globalConfig: config, PluginConfig: pluginCfg}
}

// ManuallyInstalled returns whether the plugin should be installed manually.
// If source or version is omitted, you will have to install it manually.
func (c *InstallConfig) ManuallyInstalled() bool {
	return c.Version == "" || c.Source == ""
}

// InstallPath returns an installation path from the plugin directory.
func (c *InstallConfig) InstallPath() string {
	return filepath.Join(c.Source, c.Version, fmt.Sprintf("tflint-ruleset-%s", c.Name))
}

// TagName returns a tag name that the GitHub release should meet.
// The version must not contain leading "v", as the prefix "v" is added here,
// and the release tag must be in a format similar to `v1.1.1`.
func (c *InstallConfig) TagName() string {
	return fmt.Sprintf("v%s", c.Version)
}

// AssetName returns a name that the asset contained in the release should meet.
// The name must be in a format similar to `tflint-ruleset-aws_darwin_amd64.zip`.
func (c *InstallConfig) AssetName() string {
	return fmt.Sprintf("tflint-ruleset-%s_%s_%s.zip", c.Name, runtime.GOOS, runtime.GOARCH)
}

// Install fetches the release from GitHub and puts the binary in the plugin directory.
// This installation process will automatically check the checksum of the downloaded zip file.
// Therefore, the release must always contain a checksum file.
// In addition, the release must meet the following conventions:
//
//   - The release must be tagged with a name like v1.1.1
//   - The release must contain an asset with a name like tflint-ruleset-{name}_{GOOS}_{GOARCH}.zip
//   - The zip file must contain a binary named tflint-ruleset-{name} (tflint-ruleset-{name}.exe in Windows)
//   - The release must contain a checksum file for the zip file with the name checksums.txt
//   - The checksum file must contain a sha256 hash and filename
//
// For security, you can also make sure that the checksum file is signed correctly.
// In that case, the release must additionally meet the following conventions:
//
//   - The release must contain a signature file for the checksum file with the name checksums.txt.sig
//   - The signature file must be binary OpenPGP format
func (c *InstallConfig) Install() (string, error) {
	dir, err := getPluginDir(c.globalConfig)
	if err != nil {
		return "", fmt.Errorf("Failed to get plugin dir: %w", err)
	}

	path := filepath.Join(dir, c.InstallPath()+fileExt())
	log.Printf("[DEBUG] Mkdir plugin dir: %s", filepath.Dir(path))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("Failed to mkdir to %s: %w", filepath.Dir(path), err)
	}

	assets, err := c.fetchReleaseAssets()
	if err != nil {
		return "", fmt.Errorf("Failed to fetch GitHub releases: %w", err)
	}

	log.Printf("[DEBUG] Download checksums.txt")
	checksumsFile, err := c.downloadToTempFile(assets["checksums.txt"])
	if checksumsFile != nil {
		defer os.Remove(checksumsFile.Name())
	}
	if err != nil {
		return "", fmt.Errorf("Failed to download checksums.txt: %s", err)
	}

	sigchecker := NewSignatureChecker(c)
	if sigchecker.HasSigningKey() {
		log.Printf("[DEBUG] Download checksums.txt.sig")
		signatureFile, err := c.downloadToTempFile(assets["checksums.txt.sig"])
		if signatureFile != nil {
			defer os.Remove(signatureFile.Name())
		}
		if err != nil {
			return "", fmt.Errorf("Failed to download checksums.txt.sig: %s", err)
		}

		if err := sigchecker.Verify(checksumsFile, signatureFile); err != nil {
			return "", fmt.Errorf("Failed to check checksums.txt signature: %s", err)
		}
		if _, err := checksumsFile.Seek(0, 0); err != nil {
			return "", fmt.Errorf("Failed to check checksums.txt signature: %s", err)
		}
		log.Printf("[DEBUG] Verified signature successfully")
	}

	log.Printf("[DEBUG] Download %s", c.AssetName())
	zipFile, err := c.downloadToTempFile(assets[c.AssetName()])
	if zipFile != nil {
		defer os.Remove(zipFile.Name())
	}
	if err != nil {
		return "", fmt.Errorf("Failed to download %s: %s", c.AssetName(), err)
	}

	checksummer, err := NewChecksummer(checksumsFile)
	if err != nil {
		return "", fmt.Errorf("Failed to parse checksums file: %s", err)
	}
	if err = checksummer.Verify(c.AssetName(), zipFile); err != nil {
		return "", fmt.Errorf("Failed to verify checksums: %s", err)
	}
	log.Printf("[DEBUG] Matched checksum successfully")

	if err = extractFileFromZipFile(zipFile, path); err != nil {
		return "", fmt.Errorf("Failed to extract binary from %s: %s", c.AssetName(), err)
	}

	log.Printf("[DEBUG] Installed %s successfully", path)
	return path, nil
}

// fetchReleaseAssets fetches assets from the GitHub release.
// The release is determined by the source path and tag name.
func (c *InstallConfig) fetchReleaseAssets() (map[string]*github.ReleaseAsset, error) {
	assets := map[string]*github.ReleaseAsset{}

	ctx := context.Background()

	client, err := newGitHubClient(ctx, c)
	if err != nil {
		return assets, err
	}

	release, _, err := client.Repositories.GetReleaseByTag(ctx, c.SourceOwner, c.SourceRepo, c.TagName())
	if err != nil {
		return assets, err
	}

	for _, asset := range release.Assets {
		log.Printf("[DEBUG] asset found: %s", asset.GetName())
		assets[asset.GetName()] = asset
	}
	return assets, nil
}

// downloadToTempFile download assets from GitHub to a local temp file.
// It is the caller's responsibility to delete the generated the temp file.
func (c *InstallConfig) downloadToTempFile(asset *github.ReleaseAsset) (*os.File, error) {
	if asset == nil {
		return nil, fmt.Errorf("file not found in the GitHub release. Does the release contain the file with the correct name ?")
	}

	ctx := context.Background()

	client, err := newGitHubClient(ctx, c)
	if err != nil {
		return nil, err
	}

	downloader, _, err := client.Repositories.DownloadReleaseAsset(ctx, c.SourceOwner, c.SourceRepo, asset.GetID(), http.DefaultClient)
	if err != nil {
		return nil, err
	}

	file, err := os.CreateTemp("", "tflint-download-temp-file-*")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(file, downloader); err != nil {
		return file, err
	}
	downloader.Close()
	if _, err := file.Seek(0, 0); err != nil {
		return file, err
	}

	log.Printf("[DEBUG] Downloaded to %s", file.Name())
	return file, nil
}

// getGitHubToken gets a GitHub access token from environment variables.
// Environment variables are used in the following order of priority:
//
//   - GITHUB_TOKEN_{source_host} (e.g. GITHUB_TOKEN_example_com)
//   - GITHUB_TOKEN
//
// In most cases, GITHUB_TOKEN will meet your requirements, but GITHUB_TOKEN_{source_host}
// can be useful, for example if you are hosting your plugin on GHES.
// The host name must be normalized with Punycode, and "-" can be converted to "__" and "." to "-".
func (c *InstallConfig) getGitHubToken() string {
	prefix := "GITHUB_TOKEN_"
	for _, env := range os.Environ() {
		eqIdx := strings.Index(env, "=")
		if eqIdx < 0 {
			continue
		}
		name := env[:eqIdx]
		value := env[eqIdx+1:]

		if !strings.HasPrefix(name, prefix) {
			continue
		}

		rawHost := name[len(prefix):]
		rawHost = strings.ReplaceAll(rawHost, "__", "-")
		rawHost = strings.ReplaceAll(rawHost, "_", ".")
		host, err := idna.Lookup.ToUnicode(rawHost)
		if err != nil {
			log.Printf(`[DEBUG] Failed to convert "%s" to Unicode format: %s`, rawHost, err)
			continue
		}

		if host != c.SourceHost {
			continue
		}
		log.Printf("[DEBUG] %s set, plugin requests to %s will be authenticated", name, c.SourceHost)
		return value
	}

	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		log.Printf("[DEBUG] GITHUB_TOKEN set, plugin requests to the GitHub API will be authenticated")
		return t
	}

	return ""
}

func extractFileFromZipFile(zipFile *os.File, savePath string) error {
	zipFileStat, err := zipFile.Stat()
	if err != nil {
		return err
	}
	zipReader, err := zip.NewReader(zipFile, zipFileStat.Size())
	if err != nil {
		return err
	}

	var reader io.ReadCloser
	for _, f := range zipReader.File {
		log.Printf("[DEBUG] file found in zip: %s", f.Name)
		if f.Name != filepath.Base(savePath) {
			continue
		}

		reader, err = f.Open()
		if err != nil {
			return err
		}
		break
	}
	if reader == nil {
		return fmt.Errorf("file not found. Does the zip contain %s ?", filepath.Base(savePath))
	}

	file, err := os.OpenFile(savePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(file.Name())
		return err
	}

	return nil
}

func newGitHubClient(ctx context.Context, config *InstallConfig) (*github.Client, error) {
	hc := &http.Client{
		Transport: http.DefaultTransport,
	}

	if t := config.getGitHubToken(); t != "" {
		hc = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: t,
		}))
	}

	hc.Transport = &requestLoggingTransport{hc.Transport}

	if config.SourceHost == defaultSourceHost {
		return github.NewClient(hc), nil
	}

	baseURL := fmt.Sprintf("https://%s/", config.SourceHost)
	return github.NewEnterpriseClient(baseURL, baseURL, hc)
}

func fileExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// requestLoggingTransport wraps an existing RoundTripper and prints DEBUG logs before each request
type requestLoggingTransport struct {
	http.RoundTripper
}

func (s *requestLoggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	log.Printf("[DEBUG] Request to %s", r.URL)
	return s.RoundTripper.RoundTrip(r)
}
