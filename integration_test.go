package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/formatter"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
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
			Name:    "basic",
			Command: "./tflint --format json",
			Dir:     "basic",
		},
		{
			Name:    "override",
			Command: "./tflint --format json",
			Dir:     "override",
		},
		{
			Name:    "variables",
			Command: "./tflint --format json --var-file variables.tfvars --var var=var",
			Dir:     "variables",
		},
		{
			Name:    "module",
			Command: "./tflint --format json --module --ignore-module ./ignore_module",
			Dir:     "module",
		},
		{
			Name:    "without_module_init",
			Command: "./tflint --format json",
			Dir:     "without_module_init",
		},
		{
			Name:    "arguments",
			Command: fmt.Sprintf("./tflint --format json %s", filepath.Join("dir", "template.tf")),
			Env: map[string]string{
				"TF_DATA_DIR": filepath.Join("dir", ".terraform"),
			},
			Dir: "arguments",
		},
		{
			Name:    "plugin",
			Command: "./tflint --format json --module",
			Dir:     "plugin",
		},
		{
			Name:    "jsonsyntax",
			Command: "./tflint --format json",
			Dir:     "jsonsyntax",
		},
	}

	dir, _ := os.Getwd()
	defer os.Chdir(dir)

	for _, tc := range cases {
		testDir := dir + "/integration/" + tc.Dir
		os.Chdir(testDir)

		if tc.Env != nil {
			for k, v := range tc.Env {
				os.Setenv(k, v)
			}
		}

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := cmd.NewCLI(outStream, errStream)
		args := strings.Split(tc.Command, " ")
		cli.Run(args)

		var b []byte
		var err error
		if runtime.GOOS == "windows" && IsWindowsResultExist() {
			b, err = ioutil.ReadFile("result_windows.json")
		} else {
			b, err = ioutil.ReadFile("result.json")
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

		if !cmp.Equal(got, expected) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(expected, got))
		}

		if tc.Env != nil {
			for k := range tc.Env {
				os.Unsetenv(k)
			}
		}
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
