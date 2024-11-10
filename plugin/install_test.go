package plugin

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func TestIsExperimentalModeEnabled(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want bool
	}{
		{
			name: "no env",
			want: false,
		},
		{
			name: "TFLINT_EXPERIMENTAL=true",
			envs: map[string]string{
				"TFLINT_EXPERIMENTAL": "true",
			},
			want: true,
		},
		{
			name: "TFLINT_EXPERIMENTAL=false",
			envs: map[string]string{
				"TFLINT_EXPERIMENTAL": "false",
			},
			want: false,
		},
		{
			name: "TFLINT_EXPERIMENTAL=1",
			envs: map[string]string{
				"TFLINT_EXPERIMENTAL": "1",
			},
			want: true,
		},
		{
			name: "TFLINT_EXPERIMENTAL=0",
			envs: map[string]string{
				"TFLINT_EXPERIMENTAL": "0",
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envs {
				t.Setenv(k, v)
			}
			got := IsExperimentalModeEnabled()
			if got != test.want {
				t.Errorf("want=%t, got=%t", test.want, got)
			}
		})
	}
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

func Test_Install_Keyless(t *testing.T) {
	t.Setenv("TFLINT_EXPERIMENTAL", "true")

	originalPluginRoot := PluginRoot
	PluginRoot = t.TempDir()
	originalSigningKey := builtinSigningKey
	builtinSigningKey = "" // disable built-in signing key
	defer func() {
		PluginRoot = originalPluginRoot
		builtinSigningKey = originalSigningKey
	}()

	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        "aws",
		Enabled:     true,
		Version:     "0.35.0",
		Source:      "github.com/terraform-linters/tflint-ruleset-aws",
		SourceHost:  "github.com",
		SourceOwner: "terraform-linters",
		SourceRepo:  "tflint-ruleset-aws",
	})

	// Because the built-in signing key is disabled, an error should be returned,
	// but because artifact attestation is present, no error occurs.
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

func Test_Install_Keyless_withoutAttestation(t *testing.T) {
	originalPluginRoot := PluginRoot
	PluginRoot = t.TempDir()
	originalSigningKey := builtinSigningKey
	builtinSigningKey = "" // disable built-in signing key
	defer func() {
		PluginRoot = originalPluginRoot
		builtinSigningKey = originalSigningKey
	}()

	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        "aws",
		Enabled:     true,
		Version:     "0.31.0", // This is the last version that does not support Artifact Attestation
		Source:      "github.com/terraform-linters/tflint-ruleset-aws",
		SourceHost:  "github.com",
		SourceOwner: "terraform-linters",
		SourceRepo:  "tflint-ruleset-aws",
	})

	path, err := config.Install()
	if err == nil {
		t.Fatal("config.Install() should return ErrPluginNotVerified, but did not")
	}
	if !errors.Is(err, ErrPluginNotVerified) {
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

func Test_Install_Keyless_withoutAttestation_InExperimentalMode(t *testing.T) {
	// In experimental mode, if there is no attestation, an error will occur.
	t.Setenv("TFLINT_EXPERIMENTAL", "true")

	originalPluginRoot := PluginRoot
	PluginRoot = t.TempDir()
	originalSigningKey := builtinSigningKey
	builtinSigningKey = "" // disable built-in signing key
	defer func() {
		PluginRoot = originalPluginRoot
		builtinSigningKey = originalSigningKey
	}()

	config := NewInstallConfig(tflint.EmptyConfig(), &tflint.PluginConfig{
		Name:        "aws",
		Enabled:     true,
		Version:     "0.31.0", // This is the last version that does not support Artifact Attestation
		Source:      "github.com/terraform-linters/tflint-ruleset-aws",
		SourceHost:  "github.com",
		SourceOwner: "terraform-linters",
		SourceRepo:  "tflint-ruleset-aws",
	})

	_, err := config.Install()
	if err == nil {
		t.Fatal("config.Install() should return an error, but did not")
	}
	wantErr := "Failed to download artifact attestations: GET https://api.github.com/repos/terraform-linters/tflint-ruleset-aws/attestations/sha256:2263ed2f64b535a95ab7d19ff22b366bf6b36fb84e4f7fa879f85da698a96595: 404 Not Found []"
	if err.Error() != wantErr {
		t.Fatalf("want=%s, got=%s", wantErr, err)
	}
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
