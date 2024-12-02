package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/tflint"
)

func TestIntegration(t *testing.T) {
	// Disable the bundled plugin because the `os.Executable()` is go(1) in the tests
	tflint.DisableBundledPlugin = true
	defer func() {
		tflint.DisableBundledPlugin = false
	}()

	current, _ := os.Getwd()
	dir := filepath.Join(current, "basic")

	defer func() {
		if err := os.Chdir(current); err != nil {
			t.Fatal(err)
		}
	}()
	pluginDir := t.TempDir()
	os.Setenv("TFLINT_PLUGIN_DIR", pluginDir)
	defer os.Setenv("TFLINT_PLUGIN_DIR", "")

	// Init on the current directory
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli, err := cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	cli.Run([]string{"./tflint"})
	if !strings.Contains(errStream.String(), `Plugin "aws" not found. Did you run "tflint --init"?`) {
		t.Fatalf("Expected to contain an initialization error, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--init"})
	if !strings.Contains(outStream.String(), `Installing "aws" plugin...`) {
		t.Fatalf("Expected to contain an installation log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
	if !strings.Contains(outStream.String(), `Installed "aws" (source: github.com/terraform-linters/tflint-ruleset-aws, version: 0.21.1)`) {
		t.Fatalf("Expected to contain an installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--init"})
	if !strings.Contains(outStream.String(), `Plugin "aws" is already installed`) {
		t.Fatalf("Expected to contain an already installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--version"})
	if !strings.Contains(outStream.String(), "+ ruleset.aws (0.21.1)") {
		t.Fatalf("Expected to contain an plugin version log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	// Init with --chdir
	if err := os.Chdir(current); err != nil {
		t.Fatal(err)
	}
	outStream, errStream = new(bytes.Buffer), new(bytes.Buffer)
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	cli.Run([]string{"./tflint", "--chdir", "basic", "--init"})
	if !strings.Contains(outStream.String(), `Plugin "aws" is already installed`) {
		t.Fatalf("Expected to contain an already installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--chdir", "basic", "--version"})
	if !strings.Contains(outStream.String(), "+ ruleset.aws (0.21.1)") {
		t.Fatalf("Expected to contain an plugin version log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	// Init with --recursive
	if err := os.Chdir(current); err != nil {
		t.Fatal(err)
	}
	outStream, errStream = new(bytes.Buffer), new(bytes.Buffer)
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	cli.Run([]string{"./tflint", "--recursive", "--init"})
	if !strings.Contains(outStream.String(), "Installing plugins on each working directory...") {
		t.Fatalf("Expected to contain working dir log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
	if !strings.Contains(outStream.String(), "All plugins are already installed") {
		t.Fatalf("Expected to contain alread installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	outStream, errStream = new(bytes.Buffer), new(bytes.Buffer)
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}
	cli.Run([]string{"./tflint", "--recursive", "--version"})
	if !strings.Contains(outStream.String(), "working directory: basic") {
		t.Fatalf("Expected to contain working dir log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
	if !strings.Contains(outStream.String(), "+ ruleset.aws (0.21.1)") {
		t.Fatalf("Expected to contain an plugin version log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
}
