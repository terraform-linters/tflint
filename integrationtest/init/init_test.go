package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/cmd"
)

func TestIntegration(t *testing.T) {
	current, _ := os.Getwd()
	dir := filepath.Join(current, "basic")

	defer func() {
		if err := os.Chdir(current); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	pluginDir := t.TempDir()
	os.Setenv("TFLINT_PLUGIN_DIR", pluginDir)
	defer os.Setenv("TFLINT_PLUGIN_DIR", "")

	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := cmd.NewCLI(outStream, errStream)

	cli.Run([]string{"./tflint"})
	if !strings.Contains(errStream.String(), "Plugin `aws` not found. Did you run `tflint --init`?") {
		t.Fatalf("Expected to contain an initialization error, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--init"})
	if !strings.Contains(outStream.String(), "Installing `aws` plugin...") {
		t.Fatalf("Expected to contain an installation log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
	if !strings.Contains(outStream.String(), "Installed `aws` (source: github.com/terraform-linters/tflint-ruleset-aws, version: 0.5.0)") {
		t.Fatalf("Expected to contain an installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--init"})
	if !strings.Contains(outStream.String(), "Plugin `aws` is already installed") {
		t.Fatalf("Expected to contain an already installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
}
