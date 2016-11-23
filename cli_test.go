package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/wata727/tflint/config"
)

func TestCLIRun(t *testing.T) {
	type Result struct {
		Status     int
		Output     string
		CLIOptions TestCLIOptions
	}

	cases := []struct {
		Name    string
		Command string
		Result  Result
	}{
		{
			Name:    "print version",
			Command: "./tflint --version",
			Result: Result{
				Status: ExitCodeOK,
				Output: fmt.Sprintf("TFLint version %s", Version),
			},
		},
		{
			Name:    "print version by alias",
			Command: "./tflint -v",
			Result: Result{
				Status: ExitCodeOK,
				Output: fmt.Sprintf("TFLint version %s", Version),
			},
		},
		{
			Name:    "print help",
			Command: "./tflint --help",
			Result: Result{
				Status: ExitCodeOK,
				Output: "Usage: tflint [<options>] <args>",
			},
		},
		{
			Name:    "print help by alias",
			Command: "./tflint --h",
			Result: Result{
				Status: ExitCodeOK,
				Output: "Usage: tflint [<options>] <args>",
			},
		},
		{
			Name:    "nothing options",
			Command: "./tflint",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "load single file",
			Command: "./tflint test_template.tf",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "test_template.tf",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "enable debug option",
			Command: "./tflint --debug",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          true,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "enable debug option by alias",
			Command: "./tflint -d",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          true,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "specify json format",
			Command: "./tflint --format json",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "json",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "specify json format by alias",
			Command: "./tflint -f json",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "json",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "ignore_rules",
			Command: "./tflint --ignore-rule rule1,rule2",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{"rule1": true, "rule2": true},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "ignore_modules",
			Command: "./tflint --ignore-module module1,module2",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{"module1": true, "module2": true},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "specify config gile",
			Command: "./tflint --config .tflint.example.hcl",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.example.hcl",
				},
			},
		},
		{
			Name:    "specify config gile by alias",
			Command: "./tflint -c .tflint.example.hcl",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.example.hcl",
				},
			},
		},
		{
			Name:    "enable deep check mode",
			Command: "./tflint --deep",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      true,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "set aws credentials",
			Command: "./tflint --deep --aws-access-key AWS_ACCESS_KEY_ID --aws-secret-key AWS_SECRET_ACCESS_KEY --aws-region us-east-1",
			Result: Result{
				Status: ExitCodeOK,
				Output: "",
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:     false,
						DeepCheck: true,
						AwsCredentials: map[string]string{
							"access_key": "AWS_ACCESS_KEY_ID",
							"secret_key": "AWS_SECRET_ACCESS_KEY",
							"region":     "us-east-1",
						},
						IgnoreModule: map[string]bool{},
						IgnoreRule:   map[string]bool{},
					},
					Format:     "default",
					LoadFile:   "",
					ConfigFile: ".tflint.hcl",
				},
			},
		},
	}

	for _, tc := range cases {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{outStream: outStream, errStream: errStream, testMode: true}
		args := strings.Split(tc.Command, " ")

		status := cli.Run(args)
		if status != tc.Result.Status {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", status, tc.Result.Status, tc.Name)
		}
		var output string = outStream.String()
		if status == ExitCodeError {
			output = errStream.String()
		}
		if !strings.Contains(output, tc.Result.Output) {
			t.Fatalf("Bad: %s\nExpected Contains: %s\n\ntestcase: %s", output, tc.Result.Output, tc.Name)
		}
		if !reflect.DeepEqual(cli.TestCLIOptions, tc.Result.CLIOptions) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", cli.TestCLIOptions, tc.Result.CLIOptions, tc.Name)
		}
	}
}
