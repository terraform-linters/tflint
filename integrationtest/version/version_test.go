package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/tflint"
)

func TestVersionRecursiveWithPlugins(t *testing.T) {
	// Disable the bundled plugin because os.Executable() returns go(1) in tests
	tflint.DisableBundledPlugin = true
	t.Cleanup(func() {
		tflint.DisableBundledPlugin = false
	})

	// Create test directory structure
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, ".tflint.d")
	t.Setenv("TFLINT_PLUGIN_DIR", pluginDir)

	module1 := filepath.Join(tmpDir, "module1")
	module2 := filepath.Join(tmpDir, "module2")

	if err := os.MkdirAll(module1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(module2, 0755); err != nil {
		t.Fatal(err)
	}

	// Root config: aws plugin
	rootConfig := `
plugin "aws" {
  enabled = true
  version = "0.21.1"
  source = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".tflint.hcl"), []byte(rootConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Module 1: aws plugin (duplicate)
	module1Config := `
plugin "aws" {
  enabled = true
  version = "0.21.1"
  source = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
	if err := os.WriteFile(filepath.Join(module1, ".tflint.hcl"), []byte(module1Config), 0644); err != nil {
		t.Fatal(err)
	}

	// Module 2: google plugin (different)
	module2Config := `
plugin "google" {
  enabled = true
  version = "0.21.0"
  source = "github.com/terraform-linters/tflint-ruleset-google"
}
`
	if err := os.WriteFile(filepath.Join(module2, ".tflint.hcl"), []byte(module2Config), 0644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	// First, run init to install plugins
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli, err := cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	exitCode := cli.Run([]string{"tflint", "--recursive", "--init"})
	if exitCode != cmd.ExitCodeOK {
		t.Fatalf("init failed with exit code %d\nstdout: %s\nstderr: %s", exitCode, outStream.String(), errStream.String())
	}

	// Now run version command
	outStream.Reset()
	errStream.Reset()
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFLINT_DISABLE_VERSION_CHECK", "1")
	exitCode = cli.Run([]string{"tflint", "--recursive", "--version", "--format=json"})
	if exitCode != cmd.ExitCodeOK {
		t.Fatalf("version failed with exit code %d\nstdout: %s\nstderr: %s", exitCode, outStream.String(), errStream.String())
	}

	var output cmd.VersionOutput
	if err := json.Unmarshal(outStream.Bytes(), &output); err != nil {
		t.Fatalf("failed to unmarshal JSON: %s\noutput: %s", err, outStream.String())
	}

	// Verify modules are present (3 directories: ., module1, module2)
	if len(output.Modules) != 3 {
		t.Errorf("expected 3 modules, got %d: %+v", len(output.Modules), output.Modules)
	}

	// Verify module paths
	var gotPaths []string
	for _, mod := range output.Modules {
		gotPaths = append(gotPaths, mod.Path)
	}

	expectedPaths := []string{".", "module1", "module2"}
	opts := []cmp.Option{
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}

	if diff := cmp.Diff(expectedPaths, gotPaths, opts...); diff != "" {
		t.Errorf("module paths mismatch (-want +got):\n%s", diff)
	}

	// Verify deduplicated plugins list contains both aws and google
	if len(output.Plugins) != 2 {
		t.Errorf("expected 2 deduplicated plugins (aws, google), got %d: %+v", len(output.Plugins), output.Plugins)
	}

	foundAWS := false
	foundGoogle := false
	for _, p := range output.Plugins {
		if p.Name == "ruleset.aws" && p.Version == "0.21.1" {
			foundAWS = true
		}
		if p.Name == "ruleset.google" && p.Version == "0.21.0" {
			foundGoogle = true
		}
	}

	if !foundAWS {
		t.Errorf("expected aws plugin in deduplicated list, got: %+v", output.Plugins)
	}
	if !foundGoogle {
		t.Errorf("expected google plugin in deduplicated list, got: %+v", output.Plugins)
	}

	// Verify plugins are sorted by name
	if len(output.Plugins) >= 2 {
		if output.Plugins[0].Name > output.Plugins[1].Name {
			t.Errorf("plugins should be sorted by name, got: %+v", output.Plugins)
		}
	}
}

func TestVersionNonRecursive(t *testing.T) {
	tflint.DisableBundledPlugin = true
	t.Cleanup(func() {
		tflint.DisableBundledPlugin = false
	})

	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, ".tflint.d")
	t.Setenv("TFLINT_PLUGIN_DIR", pluginDir)

	config := `
plugin "aws" {
  enabled = true
  version = "0.21.1"
  source = "github.com/terraform-linters/tflint-ruleset-aws"
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".tflint.hcl"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	// Init
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli, err := cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	exitCode := cli.Run([]string{"tflint", "--init"})
	if exitCode != cmd.ExitCodeOK {
		t.Fatalf("init failed with exit code %d\nstdout: %s\nstderr: %s", exitCode, outStream.String(), errStream.String())
	}

	// Version (non-recursive)
	outStream.Reset()
	errStream.Reset()
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFLINT_DISABLE_VERSION_CHECK", "1")
	exitCode = cli.Run([]string{"tflint", "--version", "--format=json"})
	if exitCode != cmd.ExitCodeOK {
		t.Fatalf("version failed with exit code %d\nstdout: %s\nstderr: %s", exitCode, outStream.String(), errStream.String())
	}

	var output cmd.VersionOutput
	if err := json.Unmarshal(outStream.Bytes(), &output); err != nil {
		t.Fatalf("failed to unmarshal JSON: %s\noutput: %s", err, outStream.String())
	}

	// Non-recursive mode should NOT have modules field
	if len(output.Modules) != 0 {
		t.Errorf("non-recursive mode should not have modules, got: %+v", output.Modules)
	}

	// Should have plugins field
	if len(output.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d: %+v", len(output.Plugins), output.Plugins)
	}

	if output.Plugins[0].Name != "ruleset.aws" || output.Plugins[0].Version != "0.21.1" {
		t.Errorf("expected aws plugin 0.21.1, got: %+v", output.Plugins[0])
	}
}

func TestVersionTextFormat(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, ".tflint.hcl"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmpDir)

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli, err := cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFLINT_DISABLE_VERSION_CHECK", "1")
	exitCode := cli.Run([]string{"tflint", "--version"})
	if exitCode != cmd.ExitCodeOK {
		t.Fatalf("expected exit code %d, got %d\nstderr: %s", cmd.ExitCodeOK, exitCode, errStream.String())
	}

	output := outStream.String()
	if !strings.Contains(output, "TFLint version") {
		t.Errorf("output should contain 'TFLint version', got: %s", output)
	}
}
