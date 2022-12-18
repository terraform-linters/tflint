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
			Name:    "arguments with values file",
			Command: fmt.Sprintf("./tflint --format json -c %s --var-file %s dir", filepath.Join("dir", ".tflint.hcl"), filepath.Join("dir", "subdir.tfvars")),
			Dir:     "arguments-with-values-file",
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
		{
			Name:    "path",
			Command: "./tflint --format json --module",
			Dir:     "path",
		},
		{
			Name:    "init from parent",
			Command: "./tflint --format json --module root",
			Dir:     "init-parent",
		},
		{
			Name:    "init from cwd",
			Command: "./tflint --format json --module",
			Dir:     "init-cwd/root",
		},
		{
			Name:    "enable rule which has required configuration by CLI options",
			Command: "./tflint --format json --enable-rule aws_s3_bucket_with_config_example",
			Dir:     "enable-required-config-rule-by-cli",
		},
		{
			Name:    "enable rule which does not have required configuration by CLI options",
			Command: "./tflint --format json --enable-rule aws_db_instance_with_default_config_example",
			Dir:     "enable-config-rule-by-cli",
		},
		{
			Name:    "heredoc",
			Command: "./tflint --format json",
			Dir:     "heredoc",
		},
		{
			Name:    "config parse error with HCL metadata",
			Command: "./tflint --format json",
			Dir:     "bad-config",
		},
		{
			Name:    "conditional resources",
			Command: "./tflint --format json",
			Dir:     "conditional",
		},
		{
			Name:    "dynamic blocks",
			Command: "./tflint --format json",
			Dir:     "dynblock",
		},
		{
			Name:    "unknown dynamic blocks",
			Command: "./tflint --format json",
			Dir:     "dynblock-unknown",
		},
		{
			Name:    "provider config",
			Command: "./tflint --format json",
			Dir:     "provider-config",
		},
		{
			Name:    "rule config",
			Command: "./tflint --format json",
			Dir:     "rule-config",
		},
		{
			Name:    "disabled rules",
			Command: "./tflint --format json",
			Dir:     "disabled-rules",
		},
		{
			Name:    "cty-based eval",
			Command: "./tflint --format json",
			Dir:     "cty-based-eval",
		},
		{
			Name:    "map attribute eval",
			Command: "./tflint --format json",
			Dir:     "map-attribute",
		},
		{
			Name:    "rule config with --enable-rule",
			Command: "tflint --enable-rule aws_s3_bucket_with_config_example --format json",
			Dir:     "rule-config",
		},
		{
			Name:    "rule config with --only",
			Command: "tflint --only aws_s3_bucket_with_config_example --format json",
			Dir:     "rule-config",
		},
		{
			Name:    "rule config without required attributes",
			Command: "tflint --format json",
			Dir:     "rule-required-config",
		},
		{
			Name:    "rule config without optional attributes",
			Command: "tflint --format json",
			Dir:     "rule-optional-config",
		},
		{
			Name:    "enable plugin by CLI",
			Command: "tflint --enable-plugin testing --format json",
			Dir:     "enable-plugin-by-cli",
		},
		{
			Name:    "eval on root context",
			Command: "tflint --module --format json",
			Dir:     "eval-on-root-context",
		},
		{
			Name:    "sensitve variable",
			Command: "tflint --format json",
			Dir:     "sensitive",
		},
		{
			Name:    "just attributes",
			Command: "tflint --format json",
			Dir:     "just-attributes",
		},
		{
			Name:    "incompatible host version",
			Command: "tflint --format json",
			Dir:     "incompatible-host",
		},
		{
			Name:    "expand resources/modules",
			Command: "tflint --module --format json",
			Dir:     "expand",
		},
		{
			Name:    "chdir",
			Command: "tflint --chdir dir --module --var-file from_cli.tfvars --format json",
			Dir:     "chdir",
		},
		{
			Name:    "recursive",
			Command: "tflint --recursive --format json",
			Dir:     "recursive",
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
		})
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}
