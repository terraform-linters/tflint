package tflint

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestLoadConfigGitLab(t *testing.T) {
	for name, contents := range map[string]string{
		".tflint.hcl": `plugin "example" {
  enabled = true
  source_type = "gitlab"
  source = "gitlab.example.com/group/subgroup/tflint-ruleset-example"
  version = "0.1.0"
}`,
		".tflint.json": `{"plugin":{"example":{"enabled":true,"source_type":"gitlab","source":"gitlab.example.com/group/subgroup/tflint-ruleset-example","version":"0.1.0"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			if err := fs.WriteFile(name, []byte(contents), 0o600); err != nil {
				t.Fatal(err)
			}
			config, err := LoadConfig(fs, name)
			if err != nil {
				t.Fatal(err)
			}
			plugin := config.Plugins["example"]
			if plugin.SourceType != "gitlab" || plugin.SourceHost != "gitlab.example.com" || plugin.SourceOwner != "group/subgroup" || plugin.SourceRepo != "tflint-ruleset-example" {
				t.Fatalf("unexpected plugin config: %#v", plugin)
			}
			if _, diags := plugin.Content(nil); diags.HasErrors() {
				t.Fatalf("source_type must not be passed to the ruleset: %s", diags)
			}
		})
	}
}

func TestPluginConfigGitLabValidation(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		sourceType string
		signature  string
		wantOwner  string
		wantError  string
	}{
		{name: "gitlab.com", source: "gitlab.com/group/project", sourceType: "gitlab", wantOwner: "group"},
		{name: "nested groups", source: "gitlab.example.com/group/subgroup/team/project", sourceType: "gitlab", wantOwner: "group/subgroup/team"},
		{name: "default remains GitHub", source: "github.example.com/group/project", wantOwner: "group"},
		{name: "explicit GitHub", source: "github.com/group/project", sourceType: "github", wantOwner: "group"},
		{name: "GitHub rejects nested groups", source: "github.com/group/subgroup/project", wantError: "Must be a GitHub reference"},
		{name: "unknown source type", source: "example.com/group/project", sourceType: "other", wantError: "invalid source_type"},
		{name: "GitLab attestation", source: "gitlab.com/group/project", sourceType: "gitlab", signature: "attestation", wantError: "not supported for GitLab sources"},
		{name: "GitLab PGP", source: "gitlab.com/group/project", sourceType: "gitlab", signature: "pgp", wantOwner: "group"},
	}
	for _, source := range []string{
		"gitlab.com/project", "gitlab.com//project", "gitlab.com/group/project/",
		"gitlab.com/../project", "gitlab.com/group/./project", "gitlab.com/group/../../project",
		"https://gitlab.com/group/project", "user@gitlab.com/group/project", "/group/project",
		"gitlab.com/group/project?ref=main", "gitlab.com/group/project?", "gitlab.com/group/project#",
		"gitlab.com/group/%2e%2e/project", `gitlab.com/group\subgroup/project`, "gitlab.com/group/ project",
	} {
		tests = append(tests, struct {
			name, source, sourceType, signature, wantOwner, wantError string
		}{name: source, source: source, sourceType: "gitlab", wantError: "Must be a GitLab reference"})
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &PluginConfig{Name: "example", Enabled: true, Source: test.source, SourceType: test.sourceType, Version: "0.1.0", Signature: test.signature}
			err := config.validate()
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("got %v, want error containing %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if config.SourceOwner != test.wantOwner || config.SourceRepo != "project" {
				t.Fatalf("unexpected parsed source: %#v", config)
			}
		})
	}
}
