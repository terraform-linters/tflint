package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/golang/mock/gomock"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/rules"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/terraform"
	"github.com/terraform-linters/tflint/tflint"
)

func TestCLIRun__noIssuesFound(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		LoadErr error
		Status  int
		Stdout  string
		Stderr  string
	}{
		{
			Name:    "print version",
			Command: "./tflint --version",
			Status:  ExitCodeOK,
			Stdout:  fmt.Sprintf("TFLint version %s", tflint.Version),
		},
		{
			Name:    "print help",
			Command: "./tflint --help",
			Status:  ExitCodeOK,
			Stdout:  "Application Options:",
		},
		{
			Name:    "no options",
			Command: "./tflint",
			Status:  ExitCodeOK,
			Stdout:  "",
		},
		{
			Name:    "specify format",
			Command: "./tflint --format json",
			Status:  ExitCodeOK,
			Stdout:  "[]",
		},
		{
			Name:    "`--force` option",
			Command: "./tflint --force",
			Status:  ExitCodeOK,
			Stdout:  "",
		},
		{
			Name:    "`--only` option",
			Command: "./tflint --only terraform_deprecated_interpolation",
			Status:  ExitCodeOK,
			Stdout:  "",
		},
		{
			Name:    "loading errors are occurred",
			Command: "./tflint",
			LoadErr: errors.New("Load error occurred"),
			Status:  ExitCodeError,
			Stderr:  "Load error occurred",
		},
		{
			Name:    "removed `debug` options",
			Command: "./tflint --debug",
			Status:  ExitCodeError,
			Stderr:  "`debug` option was removed in v0.8.0. Please set `TFLINT_LOG` environment variables instead",
		},
		{
			Name:    "removed `fast` option",
			Command: "./tflint --fast",
			Status:  ExitCodeError,
			Stderr:  "`fast` option was removed in v0.9.0. The `aws_instance_invalid_ami` rule is already fast enough",
		},
		{
			Name:    "removed `--error-with-issues` option",
			Command: "./tflint --error-with-issues",
			Status:  ExitCodeError,
			Stderr:  "`error-with-issues` option was removed in v0.9.0. The behavior is now default",
		},
		{
			Name:    "removed `--quiet` option",
			Command: "./tflint --quiet",
			Status:  ExitCodeError,
			Stderr:  "`quiet` option was removed in v0.11.0. The behavior is now default",
		},
		{
			Name:    "removed `--ignore-rule` option",
			Command: "./tflint --ignore-rule terraform_deprecated_interpolation",
			Status:  ExitCodeError,
			Stderr:  "`ignore-rule` option was removed in v0.12.0. Please use `--disable-rule` instead",
		},
		{
			Name:    "removed `--deep` option",
			Command: "./tflint --deep",
			Status:  ExitCodeError,
			Stderr:  "`deep` option was removed in v0.23.0. Deep checking is now a feature of the AWS plugin, so please configure the plugin instead",
		},
		{
			Name:    "removed `--aws-access-key` option",
			Command: "./tflint --aws-access-key AWS_ACCESS_KEY_ID",
			Status:  ExitCodeError,
			Stderr:  "`aws-access-key` option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			Name:    "removed `--aws-secret-key` option",
			Command: "./tflint --aws-secret-key AWS_SECRET_ACCESS_KEY",
			Status:  ExitCodeError,
			Stderr:  "`aws-secret-key` option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			Name:    "removed `--aws-profile` option",
			Command: "./tflint --aws-profile AWS_PROFILE",
			Status:  ExitCodeError,
			Stderr:  "`aws-profile` option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			Name:    "removed `--aws-creds-file` option",
			Command: "./tflint --aws-creds-file FILE",
			Status:  ExitCodeError,
			Stderr:  "`aws-creds-file` option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			Name:    "removed `--aws-region` option",
			Command: "./tflint --aws-region us-east-1",
			Status:  ExitCodeError,
			Stderr:  "`aws-region` option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			Name:    "invalid options",
			Command: "./tflint --unknown",
			Status:  ExitCodeError,
			Stderr:  "`unknown` is unknown option. Please run `tflint --help`",
		},
		{
			Name:    "invalid format",
			Command: "./tflint --format awesome",
			Status:  ExitCodeError,
			Stderr:  "Invalid value `awesome' for option",
		},
		{
			Name:    "invalid rule name",
			Command: "./tflint --enable-rule nosuchrule",
			Status:  ExitCodeError,
			Stderr:  "Rule not found: nosuchrule",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{
			outStream: outStream,
			errStream: errStream,
			testMode:  true,
		}

		loader := tflint.NewMockAbstractLoader(ctrl)
		loader.EXPECT().LoadConfig(".").Return(configs.NewEmptyConfig(), tc.LoadErr).AnyTimes()
		loader.EXPECT().Files().Return(map[string]*hcl.File{}, tc.LoadErr).AnyTimes()
		loader.EXPECT().LoadAnnotations(".").Return(map[string]tflint.Annotations{}, tc.LoadErr).AnyTimes()
		loader.EXPECT().LoadValuesFiles().Return([]terraform.InputValues{}, tc.LoadErr).AnyTimes()
		loader.EXPECT().Sources().Return(map[string][]byte{}).AnyTimes()
		cli.loader = loader

		status := cli.Run(strings.Split(tc.Command, " "))

		if status != tc.Status {
			t.Fatalf("Failed `%s`: Expected status is `%d`, but get `%d`", tc.Name, tc.Status, status)
		}
		if !strings.Contains(outStream.String(), tc.Stdout) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stdout, but get `%s`", tc.Name, tc.Stdout, outStream.String())
		}
		if tc.Stdout == "" && outStream.String() != "" {
			t.Fatalf("Failed `%s`: Expected empty in stdout, but get `%s`", tc.Name, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.Stderr) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stderr, but get `%s`", tc.Name, tc.Stderr, errStream.String())
		}
		if tc.Stderr == "" && errStream.String() != "" {
			t.Fatalf("Failed `%s`: Expected empty in stderr, but get `%s`", tc.Name, errStream.String())
		}
	}
}

type testRule struct {
	dir string
}
type errorRule struct{}

func (r *testRule) Name() string {
	return "test_rule"
}

func (r *errorRule) Name() string {
	return "error_rule"
}

func (r *testRule) Enabled() bool {
	return true
}

func (r *errorRule) Enabled() bool {
	return true
}

func (r *testRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *errorRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *testRule) Link() string {
	return ""
}

func (r *errorRule) Link() string {
	return ""
}

func (r *testRule) Check(runner *tflint.Runner) error {
	filename := "test.tf"
	if r.dir != "" {
		filename = filepath.Join(r.dir, filename)
	}

	runner.EmitIssue(
		r,
		"This is test error",
		hcl.Range{
			Filename: filename,
			Start:    hcl.Pos{Line: 1},
		},
	)
	return nil
}

func (r *errorRule) Check(runner *tflint.Runner) error {
	return errors.New("Check failed")
}

func TestCLIRun__issuesFound(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Rule    rules.Rule
		Status  int
		Stdout  string
		Stderr  string
	}{
		{
			Name:    "issues found",
			Command: "./tflint",
			Rule:    &testRule{},
			Status:  ExitCodeIssuesFound,
			Stdout:  fmt.Sprintf("%s (test_rule)", color.New(color.Bold).Sprint("This is test error")),
		},
		{
			Name:    "`--force` option",
			Command: "./tflint --force",
			Rule:    &testRule{},
			Status:  ExitCodeOK,
			Stdout:  fmt.Sprintf("%s (test_rule)", color.New(color.Bold).Sprint("This is test error")),
		},
		{
			Name:    "`--no-color` option",
			Command: "./tflint --no-color",
			Rule:    &testRule{},
			Status:  ExitCodeIssuesFound,
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "checking errors are occurred",
			Command: "./tflint",
			Rule:    &errorRule{},
			Status:  ExitCodeError,
			Stderr:  "Check failed",
		},
	}

	ctrl := gomock.NewController(t)
	originalRules := rules.DefaultRules
	defer func() {
		rules.DefaultRules = originalRules
		ctrl.Finish()
	}()

	for _, tc := range cases {
		// Mock rules
		rules.DefaultRules = []rules.Rule{tc.Rule}

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{
			outStream: outStream,
			errStream: errStream,
			testMode:  true,
		}

		loader := tflint.NewMockAbstractLoader(ctrl)
		loader.EXPECT().LoadConfig(".").Return(configs.NewEmptyConfig(), nil).AnyTimes()
		loader.EXPECT().Files().Return(map[string]*hcl.File{}, nil).AnyTimes()
		loader.EXPECT().LoadAnnotations(".").Return(map[string]tflint.Annotations{}, nil).AnyTimes()
		loader.EXPECT().LoadValuesFiles().Return([]terraform.InputValues{}, nil).AnyTimes()
		loader.EXPECT().Sources().Return(map[string][]byte{}).AnyTimes()
		cli.loader = loader

		status := cli.Run(strings.Split(tc.Command, " "))

		if status != tc.Status {
			t.Fatalf("Failed `%s`: Expected status is `%d`, but get `%d`", tc.Name, tc.Status, status)
		}
		if !strings.Contains(outStream.String(), tc.Stdout) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stdout, but get `%s`", tc.Name, tc.Stdout, outStream.String())
		}
		if tc.Stdout == "" && outStream.String() != "" {
			t.Fatalf("Failed `%s`: Expected empty in stdout, but get `%s`", tc.Name, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.Stderr) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stderr, but get `%s`", tc.Name, tc.Stderr, errStream.String())
		}
		if tc.Stderr == "" && errStream.String() != "" {
			t.Fatalf("Failed `%s`: Expected empty in stderr, but get `%s`", tc.Name, errStream.String())
		}
	}
}

func TestCLIRun__withArguments(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Dir     string
		Status  int
		Stdout  string
		Stderr  string
	}{
		{
			Name:    "no arguments",
			Command: "./tflint",
			Dir:     ".",
			Status:  ExitCodeIssuesFound,
			Stdout:  fmt.Sprintf("%s (test_rule)", color.New(color.Bold).Sprint("This is test error")),
		},
		{
			Name:    "files arguments",
			Command: "./tflint template.tf",
			Dir:     ".",
			Status:  ExitCodeOK,
			Stdout:  "",
		},
		{
			Name:    "file not found",
			Command: "./tflint not_found.tf",
			Dir:     ".",
			Status:  ExitCodeError,
			Stderr:  "Failed to load `not_found.tf`: File not found",
		},
		{
			Name:    "not Terraform configuration",
			Command: "./tflint README",
			Dir:     ".",
			Status:  ExitCodeError,
			Stderr:  "Failed to load `README`: File is not a target of Terraform",
		},
		{
			Name:    "multiple files",
			Command: "./tflint template.tf test.tf",
			Dir:     ".",
			Status:  ExitCodeIssuesFound,
			Stdout:  fmt.Sprintf("%s (test_rule)", color.New(color.Bold).Sprint("This is test error")),
		},
		{
			Name:    "directory argument",
			Command: "./tflint example",
			Dir:     "example",
			Status:  ExitCodeIssuesFound,
			Stdout:  fmt.Sprintf("%s (test_rule)", color.New(color.Bold).Sprint("This is test error")),
		},
		{
			Name:    "file under the directory",
			Command: fmt.Sprintf("./tflint %s", filepath.Join("example", "test.tf")),
			Dir:     "example",
			Status:  ExitCodeIssuesFound,
			Stdout:  fmt.Sprintf("%s (test_rule)", color.New(color.Bold).Sprint("This is test error")),
		},
		{
			Name:    "multiple directories",
			Command: "./tflint example ./",
			Dir:     "example",
			Status:  ExitCodeError,
			Stderr:  "Failed to load `example`: Multiple arguments are not allowed when passing a directory",
		},
		{
			Name:    "file and directory",
			Command: "./tflint template.tf example",
			Dir:     "example",
			Status:  ExitCodeError,
			Stderr:  "Failed to load `example`: Multiple arguments are not allowed when passing a directory",
		},
		{
			Name:    "multiple files in different directories",
			Command: fmt.Sprintf("./tflint test.tf %s", filepath.Join("example", "test.tf")),
			Dir:     "example",
			Status:  ExitCodeError,
			Stderr:  fmt.Sprintf("Failed to load `%s`: Multiple files in different directories are not allowed", filepath.Join("example", "test.tf")),
		},
	}

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(filepath.Join(currentDir, "test-fixtures", "arguments"))
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	originalRules := rules.DefaultRules

	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Fatal(err)
		}
		rules.DefaultRules = originalRules
		ctrl.Finish()
	}()

	for _, tc := range cases {
		// Mock rules
		rules.DefaultRules = []rules.Rule{&testRule{dir: tc.Dir}}

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{
			outStream: outStream,
			errStream: errStream,
			testMode:  true,
		}

		loader := tflint.NewMockAbstractLoader(ctrl)
		loader.EXPECT().LoadConfig(tc.Dir).Return(configs.NewEmptyConfig(), nil).AnyTimes()
		loader.EXPECT().Files().Return(map[string]*hcl.File{}, nil).AnyTimes()
		loader.EXPECT().LoadAnnotations(tc.Dir).Return(map[string]tflint.Annotations{}, nil).AnyTimes()
		loader.EXPECT().LoadValuesFiles().Return([]terraform.InputValues{}, nil).AnyTimes()
		loader.EXPECT().Sources().Return(map[string][]byte{}).AnyTimes()
		cli.loader = loader

		status := cli.Run(strings.Split(tc.Command, " "))

		if status != tc.Status {
			t.Fatalf("Failed `%s`: Expected status is `%d`, but get `%d`", tc.Name, tc.Status, status)
		}
		if !strings.Contains(outStream.String(), tc.Stdout) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stdout, but get `%s`", tc.Name, tc.Stdout, outStream.String())
		}
		if tc.Stdout == "" && outStream.String() != "" {
			t.Fatalf("Failed `%s`: Expected empty in stdout, but get `%s`", tc.Name, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.Stderr) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stderr, but get `%s`", tc.Name, tc.Stderr, errStream.String())
		}
		if tc.Stderr == "" && errStream.String() != "" {
			t.Fatalf("Failed `%s`: Expected empty in stderr, but get `%s`", tc.Name, errStream.String())
		}
	}
}
