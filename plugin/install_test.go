package plugin

import (
	"context"
	"os"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

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
		tc := tc

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
