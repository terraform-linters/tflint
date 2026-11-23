package langserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func TestNewHandler_lockfile(t *testing.T) {
	tflint.DisableBundledPlugin = true
	defer func() { tflint.DisableBundledPlugin = false }()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TFLINT_PLUGIN_DIR", filepath.Join(cwd, "..", "plugin", "test-fixtures", "plugins"))
	t.Chdir(t.TempDir())

	config := `
plugin "bar" {
  enabled = true
  source  = "github.com/terraform-linters/tflint-ruleset-bar"
  version = "0.1.0"
}
`
	lockfile := `
lockfile_version = 1

plugin "github.com/terraform-linters/tflint-ruleset-bar" {
  version       = "0.2.0"
  repository_id = 123
  hashes        = {}
}
`
	if err := os.WriteFile(".tflint.hcl", []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".tflint.lock.hcl", []byte(lockfile), 0o644); err != nil {
		t.Fatal(err)
	}

	_, plugin, err := NewHandler("", tflint.EmptyConfig())
	if plugin != nil {
		defer plugin.Clean()
	}
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
	want := `Plugin "bar" version 0.1.0 does not match the lockfile .tflint.lock.hcl (locked: 0.2.0)`
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error to contain %q, but got %q", want, err)
	}
}
