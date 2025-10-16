package inspection

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
			Command: "./tflint --format json --ignore-module ./ignore_module",
			Dir:     "module",
		},
		{
			Name:    "without module init",
			Command: "./tflint --format json",
			Dir:     "without_module_init",
		},
		{
			Name:    "with module init",
			Command: "./tflint --format json --call-module-type all",
			Dir:     "with_module_init",
		},
		{
			Name:    "no calling module",
			Command: "./tflint --format json --call-module-type none",
			Dir:     "no_calling_module",
		},
		{
			Name:    "plugin",
			Command: "./tflint --format json",
			Dir:     "plugin",
		},
		{
			Name:    "jsonsyntax",
			Command: "./tflint --format json",
			Dir:     "jsonsyntax",
		},
		{
			Name:    "path",
			Command: "./tflint --format json",
			Dir:     "path",
		},
		{
			Name:    "init from cwd",
			Command: "./tflint --format json",
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
			Command: "tflint --format json",
			Dir:     "eval-on-root-context",
		},
		{
			Name:    "marked values",
			Command: "tflint --format json",
			Dir:     "marked",
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
			Command: "tflint --format json",
			Dir:     "expand",
		},
		{
			Name:    "chdir",
			Command: "tflint --chdir dir --var-file from_cli.tfvars --format json",
			Dir:     "chdir",
		},
		{
			Name:    "functions",
			Command: "tflint --format json",
			Dir:     "functions",
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
