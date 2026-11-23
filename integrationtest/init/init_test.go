package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

	pluginDir := t.TempDir()
	os.Setenv("TFLINT_PLUGIN_DIR", pluginDir)
	defer os.Setenv("TFLINT_PLUGIN_DIR", "")

	// The lockfile is created next to the config file by "tflint --init"
	lockfilePath := filepath.Join(dir, ".tflint.lock.hcl")
	t.Cleanup(func() { os.Remove(lockfilePath) })

	// Init on the current directory
	t.Chdir(dir)

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
	if !strings.Contains(outStream.String(), "Created lockfile .tflint.lock.hcl") {
		t.Fatalf("Expected to contain a lockfile creation log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}
	lockfile, err := os.ReadFile(lockfilePath)
	if err != nil {
		t.Fatalf("Failed to read the lockfile: %s", err)
	}
	for _, want := range []string{
		`plugin "github.com/terraform-linters/tflint-ruleset-aws" {`,
		`version       = "0.21.1"`,
		fmt.Sprintf(`%s_%s = "h1:`, runtime.GOOS, runtime.GOARCH),
	} {
		if !strings.Contains(string(lockfile), want) {
			t.Fatalf("Expected the lockfile to contain %q, but did not: %s", want, lockfile)
		}
	}

	cli.Run([]string{"./tflint", "--init"})
	if !strings.Contains(outStream.String(), `All plugins are already installed`) {
		t.Fatalf("Expected to contain an already installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--version"})
	if !strings.Contains(outStream.String(), "+ ruleset.aws (0.21.1)") {
		t.Fatalf("Expected to contain an plugin version log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	// Init with --chdir
	t.Chdir(current)

	outStream, errStream = new(bytes.Buffer), new(bytes.Buffer)
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	cli.Run([]string{"./tflint", "--chdir", "basic", "--init"})
	if !strings.Contains(outStream.String(), `All plugins are already installed`) {
		t.Fatalf("Expected to contain an already installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	cli.Run([]string{"./tflint", "--chdir", "basic", "--version"})
	if !strings.Contains(outStream.String(), "+ ruleset.aws (0.21.1)") {
		t.Fatalf("Expected to contain an plugin version log, but did not: stdout=%s, stderr=%s", outStream, errStream)
	}

	// Init with --recursive
	t.Chdir(current)

	outStream, errStream = new(bytes.Buffer), new(bytes.Buffer)
	cli, err = cmd.NewCLI(outStream, errStream)
	if err != nil {
		t.Fatal(err)
	}

	cli.Run([]string{"./tflint", "--recursive", "--init"})
	if !strings.Contains(outStream.String(), "All plugins are already installed") {
		t.Fatalf("Expected to contain already installed log, but did not: stdout=%s, stderr=%s", outStream, errStream)
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
