package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLockfilePath(t *testing.T) {
	tests := []struct {
		config string
		want   string
	}{
		{config: "", want: ""},
		{config: ".tflint.hcl", want: ".tflint.lock.hcl"},
		{config: ".tflint.json", want: ".tflint.lock.hcl"},
		{config: "configs/aws.hcl", want: "configs/aws.lock.hcl"},
		{config: "/home/user/.tflint.hcl", want: "/home/user/.tflint.lock.hcl"},
		{config: "tflintrc", want: "tflintrc.lock.hcl"},
		{config: "config.yaml", want: "config.yaml.lock.hcl"},
	}

	for _, test := range tests {
		t.Run(test.config, func(t *testing.T) {
			got := LockfilePath(test.config)
			if got != test.want {
				t.Errorf("expected %q, but got %q", test.want, got)
			}
		})
	}
}

const (
	hashA = "h1:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hashB = "h1:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestLoadLockfile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Lockfile
		wantErr string
	}{
		{
			name: "valid",
			content: `
lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.35.0"
  repository_id = 123
  hashes = {
    linux_amd64  = "` + hashA + `"
    darwin_arm64 = "` + hashB + `"
  }
}

plugin "example.com/owner/tflint-ruleset-foo" {
  version       = "1.0.0"
  repository_id = 456
  hashes        = {}
}
`,
			want: &Lockfile{
				Version: 1,
				Plugins: map[string]*LockedPlugin{
					"github.com/terraform-linters/tflint-ruleset-aws": {
						Source:       "github.com/terraform-linters/tflint-ruleset-aws",
						Version:      "0.35.0",
						RepositoryID: 123,
						Hashes:       map[string]string{"linux_amd64": hashA, "darwin_arm64": hashB},
					},
					"example.com/owner/tflint-ruleset-foo": {
						Source:       "example.com/owner/tflint-ruleset-foo",
						Version:      "1.0.0",
						RepositoryID: 456,
						Hashes:       map[string]string{},
					},
				},
			},
		},
		{
			name: "unknown attributes and blocks are ignored",
			content: `
lockfile_version = 1
unknown = true

unknown_block {}

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.35.0"
  repository_id = 123
  hashes        = { linux_amd64 = "` + hashA + `" }
  url           = "https://example.com"
}
`,
			want: &Lockfile{
				Version: 1,
				Plugins: map[string]*LockedPlugin{
					"github.com/terraform-linters/tflint-ruleset-aws": {
						Source:       "github.com/terraform-linters/tflint-ruleset-aws",
						Version:      "0.35.0",
						RepositoryID: 123,
						Hashes:       map[string]string{"linux_amd64": hashA},
					},
				},
			},
		},
		{
			name:    "newer lockfile version",
			content: `lockfile_version = 2`,
			wantErr: "The lockfile version 2 is not supported by this version of TFLint (supported up to 1). Please upgrade TFLint.",
		},
		{
			name: "newer lockfile version with unknown structure",
			content: `
lockfile_version = 2

plugin "github.com/terraform-linters/tflint-ruleset-aws" "unknown" {
  hashes = ["h2:xxx"]
}
`,
			wantErr: "Please upgrade TFLint.",
		},
		{
			name:    "missing lockfile version",
			content: `plugin "github.com/terraform-linters/tflint-ruleset-aws" {}`,
			wantErr: `The argument "lockfile_version" is required`,
		},
		{
			name:    "invalid lockfile version",
			content: `lockfile_version = 0`,
			wantErr: "The lockfile version must be a positive integer",
		},
		{
			name: "missing attribute",
			content: `
lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version = "0.35.0"
  hashes  = {}
}
`,
			wantErr: `The argument "repository_id" is required`,
		},
		{
			name: "invalid platform",
			content: `
lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.35.0"
  repository_id = 123
  hashes        = { linux = "` + hashA + `" }
}
`,
			wantErr: `The platform "linux" must be in the form of "<os>_<arch>"`,
		},
		{
			name: "invalid hash",
			content: `
lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.35.0"
  repository_id = 123
  hashes        = { linux_amd64 = "zh:aaaa" }
}
`,
			wantErr: `must be in the form of "h1:<sha256 hex>"`,
		},
		{
			name: "duplicate entries",
			content: `
lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.35.0"
  repository_id = 123
  hashes        = {}
}

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.36.0"
  repository_id = 123
  hashes        = {}
}
`,
			wantErr: `The plugin "github.com/terraform-linters/tflint-ruleset-aws" is locked more than once`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".tflint.lock.hcl")
			if err := os.WriteFile(path, []byte(test.content), 0o644); err != nil {
				t.Fatal(err)
			}

			got, err := LoadLockfile(path)
			if test.wantErr != "" {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error to contain %q, but got %q", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if diff := cmp.Diff(test.want, got, cmpopts.IgnoreUnexported(Lockfile{})); diff != "" {
				t.Error(diff)
			}
			if got.Path() != path {
				t.Errorf("expected path %s, but got %s", path, got.Path())
			}
		})
	}
}

func TestLoadLockfile_notExist(t *testing.T) {
	got, err := LoadLockfile(filepath.Join(t.TempDir(), ".tflint.lock.hcl"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != nil {
		t.Fatalf("expected nil, but got %#v", got)
	}

	got, err = LoadLockfile("")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got != nil {
		t.Fatalf("expected nil, but got %#v", got)
	}
}

func TestLockfile_Bytes(t *testing.T) {
	lockfile := NewLockfile(".tflint.lock.hcl")
	lockfile.Plugins["github.com/terraform-linters/tflint-ruleset-google"] = &LockedPlugin{
		Source:       "github.com/terraform-linters/tflint-ruleset-google",
		Version:      "0.30.0",
		RepositoryID: 456,
		Hashes:       map[string]string{},
	}
	lockfile.Plugins["github.com/terraform-linters/tflint-ruleset-aws"] = &LockedPlugin{
		Source:       "github.com/terraform-linters/tflint-ruleset-aws",
		Version:      "0.35.0",
		RepositoryID: 123,
		Hashes:       map[string]string{"linux_amd64": hashA, "darwin_arm64": hashB},
	}

	want := `# This file is maintained automatically by "tflint --init".
# Manual edits may be lost in future updates.

lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-aws" {
  version       = "0.35.0"
  repository_id = 123
  hashes = {
    darwin_arm64 = "` + hashB + `"
    linux_amd64  = "` + hashA + `"
  }
}

plugin "github.com/terraform-linters/tflint-ruleset-google" {
  version       = "0.30.0"
  repository_id = 456
  hashes        = {}
}
`
	got := string(lockfile.Bytes())
	if diff := cmp.Diff(want, got); diff != "" {
		t.Error(diff)
	}

	// Round trip
	parsed, diags := parseLockfile([]byte(got), ".tflint.lock.hcl")
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	if diff := cmp.Diff(lockfile, parsed, cmpopts.IgnoreUnexported(Lockfile{})); diff != "" {
		t.Error(diff)
	}
}

func TestLockfile_Save(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".tflint.lock.hcl")

	lockfile := NewLockfile(path)
	lockfile.Plugins["github.com/terraform-linters/tflint-ruleset-aws"] = &LockedPlugin{
		Source:       "github.com/terraform-linters/tflint-ruleset-aws",
		Version:      "0.35.0",
		RepositoryID: 123,
		Hashes:       map[string]string{"linux_amd64": hashA},
	}

	written, err := lockfile.Save()
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected the lockfile to be written")
	}

	// Loading and saving without changes should not write the file
	loaded, err := LoadLockfile(path)
	if err != nil {
		t.Fatal(err)
	}
	written, err = loaded.Save()
	if err != nil {
		t.Fatal(err)
	}
	if written {
		t.Fatal("expected the lockfile not to be written")
	}

	// Saving with changes should write the file
	loaded.Plugins["github.com/terraform-linters/tflint-ruleset-aws"].Hashes["darwin_arm64"] = hashB
	written, err = loaded.Save()
	if err != nil {
		t.Fatal(err)
	}
	if !written {
		t.Fatal("expected the lockfile to be written")
	}
	reloaded, err := LoadLockfile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Plugins["github.com/terraform-linters/tflint-ruleset-aws"].Hashes["darwin_arm64"]; got != hashB {
		t.Fatalf("expected %s, but got %s", hashB, got)
	}
}
