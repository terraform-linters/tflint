package main

import (
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

	if err := exec.Command("go", "build", "-o", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "./sources/foo/main.go").Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("cp", "../test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "../test-fixtures/locals/.tflint.d/plugins/tflint-ruleset-foo"+fileExt()).Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("go", "build", "-o", "../test-fixtures/plugins/tflint-ruleset-bar"+fileExt(), "./sources/bar/main.go").Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("go", "build", "-o", "../../integration/plugin/.tflint.d/plugins/tflint-ruleset-example"+fileExt(), "./sources/example/main.go").Run(); err != nil {
		panic(err)
	}
}

func fileExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
