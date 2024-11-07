package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v53/github"
	"github.com/terraform-linters/tflint/tflint"
)

// Cache asset information for tests,
// as it can hit GitHub API rate limit.

var _testConfig1 *tflint.PluginConfig
var _testAssets1 map[string]*github.ReleaseAsset
var _testConfig2 *tflint.PluginConfig
var _testAssets2 map[string]*github.ReleaseAsset

func getTestAssets1(t *testing.T) (map[string]*github.ReleaseAsset, *tflint.PluginConfig) {
	if _testAssets1 == nil {
		config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
			Name:        "aws",
			Enabled:     true,
			Version:     "0.29.0",
			Source:      "github.com/terraform-linters/tflint-ruleset-aws",
			SourceHost:  "github.com",
			SourceOwner: "terraform-linters",
			SourceRepo:  "tflint-ruleset-aws",
		})
		assets, err := config.fetchReleaseAssets()
		if err != nil {
			t.Fatalf("failed to fetch asset %s: %s", config.getReleaseCacheKey(), err)
		}
		_testConfig1 = config.PluginConfig
		_testAssets1 = assets
	}
	return _testAssets1, _testConfig1
}

func getTestAssets2(t *testing.T) (map[string]*github.ReleaseAsset, *tflint.PluginConfig) {
	if _testAssets2 == nil {
		config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
			Name:        "aws",
			Enabled:     true,
			Version:     "0.30.0",
			Source:      "github.com/terraform-linters/tflint-ruleset-aws",
			SourceHost:  "github.com",
			SourceOwner: "terraform-linters",
			SourceRepo:  "tflint-ruleset-aws",
		})
		assets, err := config.fetchReleaseAssets()
		if err != nil {
			t.Fatalf("failed to fetch asset %s: %s", config.getReleaseCacheKey(), err)
		}
		_testConfig2 = config.PluginConfig
		_testAssets2 = assets
	}
	return _testAssets2, _testConfig2
}

func Test_Install(t *testing.T) {
	original := PluginRoot
	PluginRoot = t.TempDir()
	defer func() { PluginRoot = original }()

	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        "aws",
		Enabled:     true,
		Version:     "0.29.0",
		Source:      "github.com/terraform-linters/tflint-ruleset-aws",
		SourceHost:  "github.com",
		SourceOwner: "terraform-linters",
		SourceRepo:  "tflint-ruleset-aws",
	})

	path, err := config.Install()
	if err != nil {
		t.Fatalf("Failed to install: %s", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open installed binary: %s", err)
	}
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat installed binary: %s", err)
	}
	file.Close()

	expected := "tflint-ruleset-aws" + fileExt()
	if info.Name() != expected {
		t.Fatalf("Installed binary name is invalid: expected=%s, got=%s", expected, info.Name())
	}
}

func Test_InstallWithoutAPI(t *testing.T) {
	original := PluginRoot
	PluginRoot = t.TempDir()
	defer func() { PluginRoot = original }()
	pluginCacheDir := t.TempDir()

	// Scenario:
	// Precondition: enable `plugin_release_cache` and `plugin_reduce_gh_api`
	// 1. Install with broken github token. This fails.
	// 2. Install with no github token. This outputs cache information.
	// 3. Install with broken github token. THIS SUCCEEDS.
	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        "aws",
		Enabled:     true,
		Version:     "0.29.0",
		Source:      "github.com/terraform-linters/tflint-ruleset-aws",
		SourceHost:  "github.com",
		SourceOwner: "terraform-linters",
		SourceRepo:  "tflint-ruleset-aws",
	})
	config.globalConfig.PluginReleaseCache = filepath.Join(pluginCacheDir, ".plugin-release-cache.json")
	config.globalConfig.PluginReduceGhAPI = true

	t.Run("with broken token", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "broken_one")
		_, err := config.Install()
		if err == nil {
			t.Fatalf("Expects error for wrong GITHUB_TOKEN, but succeeded")
		}
	})
	t.Run("with correct token", func(t *testing.T) {
		path, err := config.Install()
		if err != nil {
			t.Fatalf("Failed to install: %s", err)
		}
		err = os.Remove(path)
		if err != nil {
			t.Fatalf("Failed to remove %s: %s", path, err)
		}
	})
	t.Run("with correct token, but with cache", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "broken_one")
		_, err := config.Install()
		if err != nil {
			t.Fatalf("Failed to install from cache: %s", err)
		}
	})
}

func TestNewGitHubClient(t *testing.T) {
	cases := []struct {
		name     string
		config   *InstallConfig
		expected string
	}{
		{
			name: "default",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "github.com",
				},
			},
			expected: "https://api.github.com/",
		},
		{
			name: "enterprise",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "github.example.com",
				},
			},
			expected: "https://github.example.com/api/v3/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := newGitHubClient(context.Background(), tc.config)
			if err != nil {
				t.Fatalf("Failed to create client: %s", err)
			}

			if client.BaseURL.String() != tc.expected {
				t.Fatalf("Unexpected API URL: want %s, got %s", tc.expected, client.BaseURL.String())
			}
		})
	}
}

func TestGetGitHubToken(t *testing.T) {
	tests := []struct {
		name   string
		config *InstallConfig
		envs   map[string]string
		want   string
	}{
		{
			name: "no token",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "github.com",
				},
			},
			want: "",
		},
		{
			name: "GITHUB_TOKEN",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "github.com",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN": "github_com_token",
			},
			want: "github_com_token",
		},
		{
			name: "GITHUB_TOKEN_example_com",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "example.com",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN_example_com": "example_com_token",
			},
			want: "example_com_token",
		},
		{
			name: "GITHUB_TOKEN and GITHUB_TOKEN_example_com",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "example.com",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN":             "github_com_token",
				"GITHUB_TOKEN_example_com": "example_com_token",
			},
			want: "example_com_token",
		},
		{
			name: "GITHUB_TOKEN_example_com and GITHUB_TOKEN_example_org",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "example.com",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN_example_com": "example_com_token",
				"GITHUB_TOKEN_example_org": "example_org_token",
			},
			want: "example_com_token",
		},
		{
			name: "GITHUB_TOKEN_{source_host} found, but source host is not matched",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "example.org",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN_example_com": "example_com_token",
			},
			want: "",
		},
		{
			name: "GITHUB_TOKEN_{source_host} and GITHUB_TOKEN found, but source host is not matched",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "example.org",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN_example_com": "example_com_token",
				"GITHUB_TOKEN":             "github_com_token",
			},
			want: "github_com_token",
		},
		{
			name: "GITHUB_TOKEN_xn--lhr645fjve.jp",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "総務省.jp",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN_xn--lhr645fjve.jp": "mic_jp_token",
			},
			want: "mic_jp_token",
		},
		{
			name: "GITHUB_TOKEN_xn____lhr645fjve_jp",
			config: &InstallConfig{
				PluginConfig: &tflint.PluginConfig{
					SourceHost: "総務省.jp",
				},
			},
			envs: map[string]string{
				"GITHUB_TOKEN_xn____lhr645fjve_jp": "mic_jp_token",
			},
			want: "mic_jp_token",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", "")
			for k, v := range test.envs {
				t.Setenv(k, v)
			}

			got := test.config.getGitHubToken()
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestStoreToReleaseCache(t *testing.T) {
	tempdir := t.TempDir()
	testAssets1, testConfig1 := getTestAssets1(t)
	testAssets2, testConfig2 := getTestAssets2(t)

	tests := []struct {
		Name               string
		PluginReleaseCache string
		Config             *tflint.PluginConfig
		Assets             map[string]*github.ReleaseAsset
		WantNotCreated     bool
		WantContent        ReleaseCache
		Err                string
	}{
		{
			Name:               "not enabled",
			PluginReleaseCache: "",
			Config:             testConfig1,
			Assets:             testAssets1,
			WantContent:        nil,
		},
		{
			Name:               "nil",
			PluginReleaseCache: filepath.Join(tempdir, ".plugin-release-cache.json"),
			Config:             testConfig1,
			Assets:             nil,
			WantNotCreated:     true,
			WantContent:        nil,
		},
		{
			Name:               "empty",
			PluginReleaseCache: filepath.Join(tempdir, ".plugin-release-cache.json"),
			Config:             testConfig1,
			Assets:             map[string]*github.ReleaseAsset{},
			WantNotCreated:     true,
			WantContent:        nil,
		},
		{
			Name:               "new cache",
			PluginReleaseCache: filepath.Join(tempdir, ".plugin-release-cache.json"),
			Config:             testConfig1,
			Assets:             testAssets1,
			WantContent: ReleaseCache{
				fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
					Type:         "github",
					GithubAssets: testAssets1,
				},
			},
		},
		{
			Name:               "append cache",
			PluginReleaseCache: filepath.Join(tempdir, ".plugin-release-cache.json"),
			Config:             testConfig2,
			Assets:             testAssets2,
			WantContent: ReleaseCache{
				fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
					Type:         "github",
					GithubAssets: testAssets1,
				},
				fmt.Sprintf("%s:v%s", testConfig2.Source, testConfig2.Version): &ReleaseCacheEntry{
					Type:         "github",
					GithubAssets: testAssets2,
				},
			},
		},
		{
			Name:               "replace cache",
			PluginReleaseCache: filepath.Join(tempdir, ".plugin-release-cache.json"),
			Config:             testConfig1,
			Assets:             testAssets2,
			WantContent: ReleaseCache{
				fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
					Type:         "github",
					GithubAssets: testAssets2,
				},
				fmt.Sprintf("%s:v%s", testConfig2.Source, testConfig2.Version): &ReleaseCacheEntry{
					Type:         "github",
					GithubAssets: testAssets2,
				},
			},
		},
		{
			Name:               "sub directory",
			PluginReleaseCache: filepath.Join(tempdir, "subdir/.plugin-release-cache.json"),
			Config:             testConfig1,
			Assets:             testAssets1,
			WantContent: ReleaseCache{
				fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
					Type:         "github",
					GithubAssets: testAssets1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			config := &InstallConfig{
				tflint.EmptyConfig(),
				test.Config,
			}
			config.globalConfig.PluginReleaseCache = test.PluginReleaseCache
			err := config.storeToReleaseCache(test.Assets)
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if test.WantNotCreated {
				_, err := os.Stat(test.PluginReleaseCache)
				if err != nil && !os.IsNotExist(err) {
					t.Errorf("unexpected error while checking file existence: %s", err)
				} else if err == nil {
					// Delete for subsequent tests
					os.Remove(test.PluginReleaseCache)
					t.Fatalf("file created; want not created")
				}
			}
			got, _ := readReleaseCache(test.PluginReleaseCache)
			if diff := cmp.Diff(test.WantContent, got); diff != "" {
				t.Errorf("contents of cache differ\n%s", diff)
			}
		})
	}
}

func TestFetchReleaseAssetsFromCache(t *testing.T) {
	tempdir := t.TempDir()
	testAssets1, testConfig1 := getTestAssets1(t)
	testAssets2, testConfig2 := getTestAssets2(t)

	cache_1and2 := filepath.Join(tempdir, ".plugin-release-cache-1and2.json")
	err := storeReleaseCache(cache_1and2, ReleaseCache{
		fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
			Type:         "github",
			GithubAssets: testAssets1,
		},
		fmt.Sprintf("%s:v%s", testConfig2.Source, testConfig2.Version): &ReleaseCacheEntry{
			Type:         "github",
			GithubAssets: testAssets2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cache_only1 := filepath.Join(tempdir, ".plugin-release-cache-only1.json")
	err = storeReleaseCache(cache_only1, ReleaseCache{
		fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
			Type:         "github",
			GithubAssets: testAssets1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cache_only1_but_not_github := filepath.Join(tempdir, ".plugin-release-cache-only1-but-not-github.json")
	err = storeReleaseCache(cache_only1_but_not_github, ReleaseCache{
		fmt.Sprintf("%s:v%s", testConfig1.Source, testConfig1.Version): &ReleaseCacheEntry{
			Type:         "not-github",
			GithubAssets: testAssets1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Name               string
		PluginReleaseCache string
		Content            ReleaseCache // nil not to create
		Config             *tflint.PluginConfig
		WantAssets         map[string]*github.ReleaseAsset
		Err                string
	}{
		{
			Name:               "not enabled",
			PluginReleaseCache: "",
			Config:             testConfig1,
			WantAssets:         nil,
		},
		{
			Name:               "hit",
			PluginReleaseCache: cache_1and2,
			Config:             testConfig1,
			WantAssets:         testAssets1,
		},
		{
			Name:               "hit another",
			PluginReleaseCache: cache_1and2,
			Config:             testConfig2,
			WantAssets:         testAssets2,
		},
		{
			Name:               "not hit",
			PluginReleaseCache: cache_only1,
			Config:             testConfig2,
			WantAssets:         nil,
		},
		{
			Name:               "not github",
			PluginReleaseCache: cache_only1_but_not_github,
			Config:             testConfig1,
			WantAssets:         nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			config := &InstallConfig{
				tflint.EmptyConfig(),
				test.Config,
			}
			config.globalConfig.PluginReleaseCache = test.PluginReleaseCache
			got, err := config.fetchReleaseAssetsFromCache()
			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				if got, want := err.Error(), test.Err; got != want {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if diff := cmp.Diff(test.WantAssets, got); diff != "" {
				t.Errorf("contents of assets differ\n%s", diff)
			}
		})
	}
}
