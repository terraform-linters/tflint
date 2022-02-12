package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/go-homedir"
)

func main() {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	if err := os.Chdir(dir); err != nil {
		panic(err)
	}

	// Package "plugin" testing
	execCommand("go", "build", "-o", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "./sources/foo/main.go")
	execCommand("cp", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "../test-fixtures/locals/.tflint.d/plugins/tflint-ruleset-foo"+fileExt())
	execCommand("go", "build", "-o", "../test-fixtures/plugins/github.com/terraform-linters/tflint-ruleset-bar/0.1.0/tflint-ruleset-bar"+fileExt(), "./sources/bar/main.go")
	execCommand("cp", "../test-fixtures/plugins/github.com/terraform-linters/tflint-ruleset-bar/0.1.0/tflint-ruleset-bar"+fileExt(), "../test-fixtures/locals/.tflint.d/plugins/github.com/terraform-linters/tflint-ruleset-bar/0.1.0/tflint-ruleset-bar"+fileExt())
	// Without .exe in Windows
	execCommand("cp", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "../test-fixtures/plugins/tflint-ruleset-baz")

	pluginDir, err := homedir.Expand("~/.tflint.d/plugins")
	if err != nil {
		panic(err)
	}

	// E2E testing
	execCommand("mkdir", "-p", pluginDir)
	execCommand("go", "build", "-o", pluginDir+"/tflint-ruleset-testing"+fileExt(), "./sources/testing/main.go")
	execCommand("go", "build", "-o", pluginDir+"/tflint-ruleset-customrulesettesting"+fileExt(), "./sources/customrulesettesting/main.go")
	execCommand("go", "build", "-o", "../../integrationtest/inspection/plugin/.tflint.d/plugins/tflint-ruleset-example"+fileExt(), "./sources/example/main.go")
}

func fileExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func execCommand(command string, args ...string) {
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("Failed to exec command: %s", stderr.String()))
	}
}
