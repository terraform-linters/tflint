package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestBundledPlugin(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Dir     string
	}{
		{
			Name:    "basic",
			Command: "tflint --format json --force",
			Dir:     "basic",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint-ruleset-aws/issues/41
			Name:    "rule config",
			Command: "tflint --format json --force",
			Dir:     "rule-config",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint/issues/1028
			Name:    "deep checking rule config",
			Command: "tflint --format json --force",
			Dir:     "deep-checking-rule-config",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint/issues/1029
			Name:    "heredoc",
			Command: "tflint --format json --force",
			Dir:     "heredoc",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint/issues/1054
			Name:    "disabled-rules",
			Command: "tflint --format json --force",
			Dir:     "disabled-rules",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint-ruleset-aws/issues/48
			Name:    "cty-based-eval",
			Command: "tflint --format json --force",
			Dir:     "cty-based-eval",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint/issues/1102
			Name:    "map-attribute",
			Command: "tflint --format json --force",
			Dir:     "map-attribute",
		},
		{
			// Regression: https://github.com/terraform-linters/tflint/issues/1103
			Name:    "rule config with --enable-rule",
			Command: "tflint --enable-rule aws_s3_bucket_name --format json --force",
			Dir:     "rule-config",
		},
		{
			Name:    "rule config with --only",
			Command: "tflint --only aws_s3_bucket_name --format json --force",
			Dir:     "rule-config",
		},
	}

	dir, _ := os.Getwd()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testDir := filepath.Join(dir, tc.Dir)

			t.Cleanup(func() {
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
			})

			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			args := strings.Split(tc.Command, " ")
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
				fmt.Fprintf(os.Stderr, "Failed to exec command (%s)\n", err)
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
			if !cmp.Equal(got, expected, opts...) {
				t.Fatalf("diff=%s", cmp.Diff(expected, got))
			}
		})
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
