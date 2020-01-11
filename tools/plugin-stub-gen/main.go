package main

import (
	"os/exec"
	"runtime"
)

func main() {
	if err := exec.Command("go", "build", "-o", "../plugin/test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "plugin-stub-gen/sources/foo/main.go").Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("cp", "../plugin/test-fixtures/plugins/tflint-ruleset-foo"+fileExt(), "../plugin/test-fixtures/locals/.tflint.d/plugins/tflint-ruleset-foo"+fileExt()).Run(); err != nil {
		panic(err)
	}
	if err := exec.Command("go", "build", "-o", "../plugin/test-fixtures/plugins/tflint-ruleset-bar"+fileExt(), "plugin-stub-gen/sources/bar/main.go").Run(); err != nil {
		panic(err)
	}
}

func fileExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
