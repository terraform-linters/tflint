package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/terraform-linters/tflint/formatter"
)

func TestIntegration(t *testing.T) {
	tests := []struct {
		name    string
		command string
		dir     string
	}{
		{
			name:    "empty module",
			command: "tflint --format json --force",
			dir:     "empty",
		},
		{
			name:    "basic",
			command: "tflint --format json --force",
			dir:     "basic",
		},
		{
			name:    "disable bundled plugin",
			command: "tflint --format json --force",
			dir:     "disable",
		},
		{
			name:    "with config",
			command: "tflint --format json --force",
			dir:     "with_config",
		},
		{
			name:    "disabled_by_default",
			command: "tflint --format json --force",
			dir:     "disabled_by_default",
		},
		{
			name:    "only",
			command: "tflint --format json --force --only terraform_unused_declarations",
			dir:     "only",
		},
	}

	dir, _ := os.Getwd()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDir := filepath.Join(dir, test.dir)

			t.Cleanup(func() {
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
			})

			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			args := strings.Split(test.command, " ")
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("tflint.exe", args[1:]...)
			} else {
				cmd = exec.Command("tflint", args[1:]...)
			}
			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cmd.Stdout = outStream
			cmd.Stderr = errStream

			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to exec command: %s", err)
			}

			var b []byte
			var err error
			if runtime.GOOS == "windows" && IsWindowsResultExist() {
				b, err = os.ReadFile(filepath.Join(testDir, "result_windows.json"))
			} else {
				b, err = os.ReadFile(filepath.Join(testDir, "result.json"))
			}
			if err != nil {
				t.Fatal(err)
			}

			var expected *formatter.JSONOutput
			if err := json.Unmarshal(b, &expected); err != nil {
				t.Fatal(err)
			}

			var got *formatter.JSONOutput
			if err := json.Unmarshal(outStream.Bytes(), &got); err != nil {
				t.Fatal(err)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(formatter.JSONRule{}, "Link"),
			}
			if diff := cmp.Diff(got, expected, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
