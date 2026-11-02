package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/tflint"
)

func TestRecursiveDirectoryIsolation(t *testing.T) {
	tflint.DisableBundledPlugin = true
	defer func() { tflint.DisableBundledPlugin = false }()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Relative diagnostic paths require the fixtures and invocation directory
	// to share a Windows volume; the system temporary directory may not.
	root, err := os.MkdirTemp(originalWd, "recursive-directories-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	for i := range 16 {
		name := fmt.Sprintf("module-%02d", i)
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		for file, contents := range map[string]string{
			".tflint.hcl": `plugin "testing" { enabled = true }`,
			"main.tf":     `resource "aws_instance" "example" { instance_type = file("value.txt") }`,
			"value.txt":   name,
		} {
			if err := os.WriteFile(filepath.Join(dir, file), []byte(contents), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Config-only directories must not start a ruleset or require its installation.
	empty := filepath.Join(root, "empty")
	if err := os.MkdirAll(empty, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(empty, ".tflint.hcl"), []byte(`plugin "missing" { enabled = true }`), 0644); err != nil {
		t.Fatal(err)
	}
	out, errs := new(bytes.Buffer), new(bytes.Buffer)
	cli, err := cmd.NewCLI(out, errs)
	if err != nil {
		t.Fatal(err)
	}
	status := cli.Run([]string{"tflint", "--recursive", "--chdir=" + root, "--max-workers=4", "--format=json", "--only=aws_instance_example_type"})
	if status != cmd.ExitCodeIssuesFound {
		t.Fatalf("exit %d: %s %s", status, out, errs)
	}
	var result formatter.JSONOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Errors) != 0 || len(result.Issues) != 16 {
		t.Fatalf("unexpected result: %s", out)
	}
	seen := map[string]bool{}
	for _, issue := range result.Issues {
		name := filepath.Base(filepath.Dir(issue.Range.Filename))
		if issue.Message != "instance type is "+name {
			t.Errorf("directory %s read another module's value: %s", name, issue.Message)
		}
		if seen[name] {
			t.Errorf("duplicate module %s", name)
		}
		seen[name] = true
	}
	wd, err := os.Getwd()
	if err != nil || wd != originalWd {
		t.Fatalf("working directory changed: %s, %v", wd, err)
	}
}
