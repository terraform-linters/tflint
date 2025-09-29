package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/terraform-linters/tflint/cmd"
	"github.com/terraform-linters/tflint/tflint"
)

func TestIntegration(t *testing.T) {
	// Disable the bundled plugin because the `os.Executable()` is go(1) in the tests
	tflint.DisableBundledPlugin = true
	defer func() {
		tflint.DisableBundledPlugin = false
	}()

	tests := []struct {
		name    string
		command string
		dir     string
		status  int
		stdout  string
		stderr  string
	}{
		{
			name:    "print version",
			command: "./tflint --version",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  fmt.Sprintf("TFLint version %s", tflint.Version),
		},
		{
			name:    "print help",
			command: "./tflint --help",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "Application Options:",
		},
		{
			name:    "no options",
			command: "./tflint",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "--format option",
			command: "./tflint --format json",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "[]",
		},
		{
			name:    "format config",
			command: "./tflint",
			dir:     "format_config",
			status:  cmd.ExitCodeOK,
			stdout: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle></checkstyle>`,
		},
		{
			name:    "--format + config",
			command: "./tflint --format json",
			dir:     "format_config",
			status:  cmd.ExitCodeOK,
			stdout:  "[]",
		},
		{
			name:    "JSON format config",
			command: "./tflint",
			dir:     "json_config",
			status:  cmd.ExitCodeOK,
			stdout: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle></checkstyle>`,
		},
		{
			name:    "HCL precedence over JSON config",
			command: "./tflint",
			dir:     "hcl_json_precedence",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "`--force` option with no issues",
			command: "./tflint --force",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "`--minimum-failure-severity` option with no issues",
			command: "./tflint --minimum-failure-severity=notice",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "`--only` option",
			command: "./tflint --only aws_instance_example_type",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "loading errors are occurred",
			command: "./tflint",
			dir:     "load_errors",
			status:  cmd.ExitCodeError,
			stderr:  "Failed to load configurations;",
		},
		{
			name:    "removed --debug options",
			command: "./tflint --debug",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--debug option was removed in v0.8.0. Please set TFLINT_LOG environment variables instead",
		},
		{
			name:    "removed --error-with-issues option",
			command: "./tflint --error-with-issues",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--error-with-issues option was removed in v0.9.0. The behavior is now default",
		},
		{
			name:    "removed --quiet option",
			command: "./tflint --quiet",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--quiet option was removed in v0.11.0. The behavior is now default",
		},
		{
			name:    "removed --ignore-rule option",
			command: "./tflint --ignore-rule aws_instance_example_type",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--ignore-rule option was removed in v0.12.0. Please use --disable-rule instead",
		},
		{
			name:    "removed --deep option",
			command: "./tflint --deep",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--deep option was removed in v0.23.0. Deep checking is now a feature of the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-access-key option",
			command: "./tflint --aws-access-key AWS_ACCESS_KEY_ID",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-access-key option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-secret-key option",
			command: "./tflint --aws-secret-key AWS_SECRET_ACCESS_KEY",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-secret-key option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-profile option",
			command: "./tflint --aws-profile AWS_PROFILE",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-profile option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-creds-file option",
			command: "./tflint --aws-creds-file FILE",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-creds-file option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-region option",
			command: "./tflint --aws-region us-east-1",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-region option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --loglevel option",
			command: "./tflint --loglevel debug",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--loglevel option was removed in v0.40.0. Please set TFLINT_LOG environment variables instead",
		},
		{
			name:    "removed --module option",
			command: "./tflint --module",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--module option was removed in v0.54.0. Use --call-module-type=all instead",
		},
		{
			name:    "removed --no-module option",
			command: "./tflint --no-module",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--no-module option was removed in v0.54.0. Use --call-module-type=none instead",
		},
		{
			name:    "invalid options",
			command: "./tflint --unknown",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  `--unknown is unknown option. Please run "tflint --help"`,
		},
		{
			name:    "invalid format",
			command: "./tflint --format awesome",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "Invalid value `awesome' for option",
		},
		{
			name:    "invalid rule name",
			command: "./tflint --enable-rule nosuchrule",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "Rule not found: nosuchrule",
		},
		{
			name:    "issues found",
			command: "./tflint",
			dir:     "issues_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--force option with issues",
			command: "./tflint --force",
			dir:     "issues_found",
			status:  cmd.ExitCodeOK,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--minimum-failure-severity option with warning issues and minimum-failure-severity notice",
			command: "./tflint --minimum-failure-severity=notice",
			dir:     "warnings_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_s3_bucket_with_config_example)", color.New(color.Bold).Sprint("bucket name is test, config=bucket")),
		},
		{
			name:    "--minimum-failure-severity option with warning issues and minimum-failure-severity warning",
			command: "./tflint --minimum-failure-severity=warning",
			dir:     "warnings_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_s3_bucket_with_config_example)", color.New(color.Bold).Sprint("bucket name is test, config=bucket")),
		},
		{
			name:    "--minimum-failure-severity option with warning issues and minimum-failure-severity error",
			command: "./tflint --minimum-failure-severity=error",
			dir:     "warnings_found",
			status:  cmd.ExitCodeOK,
			stdout:  fmt.Sprintf("%s (aws_s3_bucket_with_config_example)", color.New(color.Bold).Sprint("bucket name is test, config=bucket")),
		},
		{
			name:    "--no-color option",
			command: "./tflint --no-color",
			dir:     "issues_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  "instance type is t2.micro (aws_instance_example_type)",
		},
		{
			name:    "checking errors are occurred",
			command: "./tflint",
			dir:     "check_errors",
			status:  cmd.ExitCodeError,
			stderr:  `failed to check "aws_cloudformation_stack_error" rule: an error occurred in Check`,
		},
		{
			name:    "files arguments",
			command: "./tflint empty.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "file not found",
			command: "./tflint not_found.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "not Terraform configuration",
			command: "./tflint README.md",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "multiple files",
			command: "./tflint empty.tf main.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "directory argument",
			command: "./tflint subdir",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "file under the directory",
			command: fmt.Sprintf("./tflint %s", filepath.Join("subdir", "main.tf")),
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "multiple directories",
			command: "./tflint subdir ./",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "file and directory",
			command: "./tflint main.tf subdir",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "multiple files in different directories",
			command: fmt.Sprintf("./tflint main.tf %s", filepath.Join("subdir", "main.tf")),
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--filter",
			command: "./tflint --filter=empty.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeOK,
			stdout:  "", // main.tf is ignored
		},
		{
			name:    "--filter with multiple files",
			command: "./tflint --filter=empty.tf --filter=main.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeIssuesFound,
			// main.tf is not ignored
			stdout: fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--filter with glob (files found)",
			command: "./tflint --filter=*.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--filter with glob (files not found)",
			command: "./tflint --filter=*_generated.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "--chdir",
			command: "./tflint --chdir=subdir",
			dir:     "chdir",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is m5.2xlarge")),
		},
		{
			name:    "--chdir and file argument",
			command: "./tflint --chdir=subdir main.tf",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and directory argument",
			command: "./tflint --chdir=subdir ../",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and the current directory argument",
			command: "./tflint --chdir=subdir .",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and file under the directory argument",
			command: fmt.Sprintf("./tflint --chdir=subdir %s", filepath.Join("nested", "main.tf")),
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and --filter",
			command: "./tflint --chdir=subdir --filter=main.tf",
			dir:     "chdir",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is m5.2xlarge")),
		},
		{
			name:    "--chdir and format config",
			command: "./tflint --chdir=subdir", // Apply config in subdir
			dir:     "chdir_format",
			status:  cmd.ExitCodeOK,
			stdout:  "[]",
		},
		{
			name:    "invalid max workers",
			command: "./tflint --max-workers=0",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  `Max workers should be greater than 0`,
		},
	}

	dir, _ := os.Getwd()
	defaultNoColor := color.NoColor

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDir := filepath.Join(dir, test.dir)
			t.Cleanup(func() {
				color.NoColor = defaultNoColor
			})
			t.Chdir(testDir)

			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cli, err := cmd.NewCLI(outStream, errStream)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Split(test.command, " ")

			got := cli.Run(args)

			if got != test.status {
				t.Errorf("expected status is %d, but got %d", test.status, got)
			}
			if !strings.Contains(outStream.String(), test.stdout) || (test.stdout == "" && outStream.String() != "") {
				t.Errorf("stdout did not contain expected\n\texpected: %s\n\tgot: %s", test.stdout, outStream.String())
			}
			if !strings.Contains(errStream.String(), test.stderr) || (test.stderr == "" && errStream.String() != "") {
				t.Errorf("stderr did not contain expected\n\texpected: %s\n\tgot: %s", test.stderr, errStream.String())
			}
		})
	}
}
