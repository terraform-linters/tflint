package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/tflint"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func TestIntegration(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Env     map[string]string
		Dir     string
	}{
		{
			Name:    "simple fix",
			Command: "./tflint --format json --fix",
			Dir:     "simple",
		},
		{
			Name:    "multiple fix in a file",
			Command: "./tflint --format json --fix",
			Dir:     "multiple_fix",
		},
		{
			Name:    "ignore by annotation",
			Command: "./tflint --format json --fix",
			Dir:     "ignore_by_annotation",
		},
		{
			Name:    "multiple fix by multiple rules",
			Command: "./tflint --format json --fix",
			Dir:     "fix_by_multiple_rules",
		},
		{
			Name:    "conflict fix by multiple rules",
			Command: "./tflint --format json --fix",
			Dir:     "conflict_fix",
		},
		{
			Name:    "fix in multiple files",
			Command: "./tflint --format json --fix",
			Dir:     "multiple_files",
		},
		{
			Name:    "calling modules",
			Command: "./tflint --format json --fix",
			Dir:     "module",
		},
		{
			Name:    "--chdir",
			Command: "./tflint --chdir=dir --format json --fix",
			Dir:     "chdir",
		},
		{
			Name:    "--chdir with conflict",
			Command: "./tflint --chdir=dir --format json --fix",
			Dir:     "chdir_with_conflict",
		},
		{
			Name:    "--filter",
			Command: "./tflint --format json --fix --filter=main.tf",
			Dir:     "filter",
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
			t.Chdir(testDir)

			tfFiles := map[string][]byte{}
			err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".tf") {
					sources, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					tfFiles[path] = sources
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				// restore original files
				for path := range tfFiles {
					if err := os.WriteFile(path, tfFiles[path], 0644); err != nil {
						t.Fatal(err)
					}
				}
			}()

			resultFile := "result.json"
			if runtime.GOOS == "windows" && IsWindowsResultExist() {
				resultFile = "result_windows.json"
			}

			if tc.Env != nil {
				for k, v := range tc.Env {
					t.Setenv(k, v)
				}
			}

			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cli, err := cmd.NewCLI(outStream, errStream)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Split(tc.Command, " ")

			cli.Run(args)

			b, err := os.ReadFile(filepath.Join(testDir, resultFile))
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

			if diff := cmp.Diff(got, expected); diff != "" {
				t.Fatal(diff)
			}

			// test autofixed files
			for path := range tfFiles {
				_, err := os.Stat(path + ".fixed")
				if os.IsNotExist(err) {
					// should be unchanged
					got, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if diff := cmp.Diff(string(got), string(tfFiles[path])); diff != "" {
						t.Fatal(diff)
					}
				} else if err == nil {
					// should be changed
					got, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					want, err := os.ReadFile(path + ".fixed")
					if err != nil {
						t.Fatal(err)
					}
					if diff := cmp.Diff(string(got), string(want)); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}
		})
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
