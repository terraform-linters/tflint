package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/rules"
	"github.com/wata727/tflint/tflint"
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
			Stdout:  fmt.Sprintf("TFLint version %s", project.Version),
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
			Stdout:  "Awesome! Your code is following the best practices :)",
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
			Stdout:  "Awesome! Your code is following the best practices :)",
		},
		{
			Name:    "`--quiet` option",
			Command: "./tflint --quiet",
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
		loader.EXPECT().LoadAnnotations(".").Return(map[string]tflint.Annotations{}, tc.LoadErr).AnyTimes()
		loader.EXPECT().LoadValuesFiles().Return([]terraform.InputValues{}, tc.LoadErr).AnyTimes()
		cli.loader = loader

		status := cli.Run(strings.Split(tc.Command, " "))

		if status != tc.Status {
			t.Fatalf("Failed `%s`: Expected status is `%d`, but get `%d`", tc.Name, tc.Status, status)
		}
		if !strings.Contains(outStream.String(), tc.Stdout) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stdout, but get `%s`", tc.Name, tc.Stdout, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.Stderr) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stderr, but get `%s`", tc.Name, tc.Stderr, errStream.String())
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

func (r *testRule) Type() string {
	return issue.ERROR
}
func (r *errorRule) Type() string {
	return issue.ERROR
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
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "`--force` option",
			Command: "./tflint --force",
			Rule:    &testRule{},
			Status:  ExitCodeOK,
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "`--quiet` option",
			Command: "./tflint --quiet",
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
		loader.EXPECT().LoadAnnotations(".").Return(map[string]tflint.Annotations{}, nil).AnyTimes()
		loader.EXPECT().LoadValuesFiles().Return([]terraform.InputValues{}, nil).AnyTimes()
		cli.loader = loader

		status := cli.Run(strings.Split(tc.Command, " "))

		if status != tc.Status {
			t.Fatalf("Failed `%s`: Expected status is `%d`, but get `%d`", tc.Name, tc.Status, status)
		}
		if !strings.Contains(outStream.String(), tc.Stdout) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stdout, but get `%s`", tc.Name, tc.Stdout, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.Stderr) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stderr, but get `%s`", tc.Name, tc.Stderr, errStream.String())
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
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "files arguments",
			Command: "./tflint template.tf",
			Dir:     ".",
			Status:  ExitCodeOK,
			Stdout:  "Awesome! Your code is following the best practices :)",
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
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "directory argument",
			Command: "./tflint example",
			Dir:     "example",
			Status:  ExitCodeIssuesFound,
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "file under the directory",
			Command: fmt.Sprintf("./tflint %s", filepath.Join("example", "test.tf")),
			Dir:     "example",
			Status:  ExitCodeIssuesFound,
			Stdout:  "This is test error (test_rule)",
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
		os.Chdir(currentDir)
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
		loader.EXPECT().LoadAnnotations(tc.Dir).Return(map[string]tflint.Annotations{}, nil).AnyTimes()
		loader.EXPECT().LoadValuesFiles().Return([]terraform.InputValues{}, nil).AnyTimes()
		cli.loader = loader

		status := cli.Run(strings.Split(tc.Command, " "))

		if status != tc.Status {
			t.Fatalf("Failed `%s`: Expected status is `%d`, but get `%d`", tc.Name, tc.Status, status)
		}
		if !strings.Contains(outStream.String(), tc.Stdout) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stdout, but get `%s`", tc.Name, tc.Stdout, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.Stderr) {
			t.Fatalf("Failed `%s`: Expected to contain `%s` in stderr, but get `%s`", tc.Name, tc.Stderr, errStream.String())
		}
	}
}
