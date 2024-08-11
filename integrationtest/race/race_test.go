package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/tflint"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

type meta struct {
	Version string
}

func TestIntegration(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Dir     string
	}{
		{
			// @see https://github.com/terraform-linters/tflint/issues/2094
			Name:    "eval locals on the root context in parallel runners",
			Command: "tflint --format json",
			Dir:     "eval_locals_on_root_ctx",
		},
	}

	// Disable the bundled plugin because the `os.Executable()` is go(1) in the tests
	tflint.DisableBundledPlugin = true
	defer func() {
		tflint.DisableBundledPlugin = false
	}()

	dir, _ := os.Getwd()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testDir := filepath.Join(dir, tc.Dir)

			defer func() {
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
			}()
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cli, err := cmd.NewCLI(outStream, errStream)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Split(tc.Command, " ")

			cli.Run(args)

			rawWant, err := readResultFile(testDir)
			if err != nil {
				t.Fatal(err)
			}
			var want *formatter.JSONOutput
			if err := json.Unmarshal(rawWant, &want); err != nil {
				t.Fatal(err)
			}

			var got *formatter.JSONOutput
			if err := json.Unmarshal(outStream.Bytes(), &got); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func readResultFile(dir string) ([]byte, error) {
	resultFile := "result.json"
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(filepath.Join(dir, "result_windows.json")); !os.IsNotExist(err) {
			resultFile = "result_windows.json"
		}
	}
	if _, err := os.Stat(fmt.Sprintf("%s.tmpl", resultFile)); !os.IsNotExist(err) {
		resultFile = fmt.Sprintf("%s.tmpl", resultFile)
	}

	if !strings.HasSuffix(resultFile, ".tmpl") {
		return os.ReadFile(filepath.Join(dir, resultFile))
	}

	want := new(bytes.Buffer)
	tmpl := template.Must(template.ParseFiles(filepath.Join(dir, resultFile)))
	if err := tmpl.Execute(want, meta{Version: tflint.Version.String()}); err != nil {
		return nil, err
	}
	return want.Bytes(), nil
}
