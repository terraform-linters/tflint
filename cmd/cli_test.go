package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform/configs"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
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
			Stdout:  fmt.Sprintf("TFLint version %s", Version),
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
			Name:    "`--error-with-issues` option",
			Command: "./tflint --error-with-issues",
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
			Name:    "invalid options",
			Command: "./tflint --debug",
			Status:  ExitCodeError,
			Stderr:  "`debug` is unknown option. Please run `tflint --help`",
		},
		{
			Name:    "invalid format",
			Command: "./tflint --format awesome",
			Status:  ExitCodeError,
			Stderr:  "Invalid value `awesome' for option",
		},
		{
			Name:    "invalid arguments",
			Command: "./tflint template.tf",
			Status:  ExitCodeError,
			Stderr:  "Too many arguments. TFLint doesn't accept the file argument",
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

		loader := mock.NewMockAbstractLoader(ctrl)
		loader.EXPECT().LoadConfig().Return(configs.NewEmptyConfig(), tc.LoadErr).AnyTimes()
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

type testRule struct{}
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

func (r *testRule) Check(runner *tflint.Runner) error {
	runner.Issues = append(runner.Issues, &issue.Issue{
		Detector: r.Name(),
		Type:     issue.ERROR,
		Message:  "This is test error",
		Line:     1,
		File:     "test.tf",
	})
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
			Status:  ExitCodeOK,
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "`--error-with-issues` option",
			Command: "./tflint --error-with-issues",
			Rule:    &testRule{},
			Status:  ExitCodeIssuesFound,
			Stdout:  "This is test error (test_rule)",
		},
		{
			Name:    "`--quiet` option",
			Command: "./tflint --quiet",
			Rule:    &testRule{},
			Status:  ExitCodeOK,
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

		loader := mock.NewMockAbstractLoader(ctrl)
		loader.EXPECT().LoadConfig().Return(configs.NewEmptyConfig(), nil).AnyTimes()
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
