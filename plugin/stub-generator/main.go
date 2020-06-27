package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	if err := os.Chdir(dir); err != nil {
		panic(err)
	}

	execCommand("go", "build", "-o", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "./sources/foo/main.go")
	execCommand("cp", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "../test-fixtures/locals/.tflint.d/plugins/tflint-ruleset-foo"+fileExt())
	execCommand("go", "build", "-o", "../test-fixtures/plugins/tflint-ruleset-bar"+fileExt(), "./sources/bar/main.go")
	execCommand("go", "build", "-o", "../../integration/plugin/.tflint.d/plugins/tflint-ruleset-example"+fileExt(), "./sources/example/main.go")
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
