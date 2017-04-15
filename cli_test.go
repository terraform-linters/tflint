package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/detector"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/loader"
	"github.com/wata727/tflint/mock"
	"github.com/wata727/tflint/printer"
)

func TestCLIRun(t *testing.T) {
	type Result struct {
		Status     int
		Stdout     string
		Stderr     string
		CLIOptions TestCLIOptions
	}
	var loaderDefaultBehavior = func(ctrl *gomock.Controller) loader.LoaderIF {
		loader := mock.NewMockLoaderIF(ctrl)
		loader.EXPECT().LoadState()
		loader.EXPECT().LoadTFVars([]string{"terraform.tfvars"})
		loader.EXPECT().LoadAllTemplate(".").Return(nil)
		return loader
	}
	var detectorNoErrorNoIssuesBehavior = func(ctrl *gomock.Controller) detector.DetectorIF {
		detector := mock.NewMockDetectorIF(ctrl)
		detector.EXPECT().Detect().Return([]*issue.Issue{})
		detector.EXPECT().HasError().Return(false)
		return detector
	}
	var printerNoIssuesDefaultBehaviour = func(ctrl *gomock.Controller) printer.PrinterIF {
		printer := mock.NewMockPrinterIF(ctrl)
		printer.EXPECT().Print([]*issue.Issue{}, "default")
		return printer
	}
	defaultCLIOptions := TestCLIOptions{
		Config: &config.Config{
			Debug:          false,
			DeepCheck:      false,
			AwsCredentials: map[string]string{},
			IgnoreModule:   map[string]bool{},
			IgnoreRule:     map[string]bool{},
			Varfile:        []string{"terraform.tfvars"},
		},
		ConfigFile: ".tflint.hcl",
	}

	cases := []struct {
		Name              string
		Command           string
		LoaderGenerator   func(ctrl *gomock.Controller) loader.LoaderIF
		DetectorGenerator func(ctrl *gomock.Controller) detector.DetectorIF
		PrinterGenerator  func(ctrl *gomock.Controller) printer.PrinterIF
		Result            Result
	}{
		{
			Name:              "print version",
			Command:           "./tflint --version",
			LoaderGenerator:   func(ctrl *gomock.Controller) loader.LoaderIF { return mock.NewMockLoaderIF(ctrl) },
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status: ExitCodeOK,
				Stdout: fmt.Sprintf("TFLint version %s", Version),
			},
		},
		{
			Name:              "print help",
			Command:           "./tflint --help",
			LoaderGenerator:   func(ctrl *gomock.Controller) loader.LoaderIF { return mock.NewMockLoaderIF(ctrl) },
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status: ExitCodeOK,
				Stdout: "Application Options:",
			},
		},
		{
			Name:              "nothing options",
			Command:           "./tflint",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:            "nothing options when issues found",
			Command:         "./tflint",
			LoaderGenerator: loaderDefaultBehavior,
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF {
				detector := mock.NewMockDetectorIF(ctrl)
				detector.EXPECT().Detect().Return([]*issue.Issue{
					&issue.Issue{
						Type:    "TEST",
						Message: "this is test method",
						Line:    1,
						File:    "",
					},
				})
				detector.EXPECT().HasError().Return(false)
				return detector
			},
			PrinterGenerator: func(ctrl *gomock.Controller) printer.PrinterIF {
				printer := mock.NewMockPrinterIF(ctrl)
				printer.EXPECT().Print([]*issue.Issue{
					&issue.Issue{
						Type:    "TEST",
						Message: "this is test method",
						Line:    1,
						File:    "",
					},
				}, "default")
				return printer
			},
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:    "nothing options when occurred loading error",
			Command: "./tflint",
			LoaderGenerator: func(ctrl *gomock.Controller) loader.LoaderIF {
				loader := mock.NewMockLoaderIF(ctrl)
				loader.EXPECT().LoadState()
				loader.EXPECT().LoadTFVars([]string{"terraform.tfvars"})
				loader.EXPECT().LoadAllTemplate(".").Return(errors.New("loading error!"))
				return loader
			},
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status:     ExitCodeError,
				Stderr:     "loading error!",
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:            "nothing options when occurred detecting error",
			Command:         "./tflint",
			LoaderGenerator: loaderDefaultBehavior,
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF {
				detector := mock.NewMockDetectorIF(ctrl)
				detector.EXPECT().Detect().Return([]*issue.Issue{})
				detector.EXPECT().HasError().Return(true)
				return detector
			},
			PrinterGenerator: func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status:     ExitCodeError,
				Stderr:     "error occurred in detecting",
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:    "load single file",
			Command: "./tflint test_template.tf",
			LoaderGenerator: func(ctrl *gomock.Controller) loader.LoaderIF {
				loader := mock.NewMockLoaderIF(ctrl)
				loader.EXPECT().LoadState()
				loader.EXPECT().LoadTFVars([]string{"terraform.tfvars"})
				loader.EXPECT().LoadTemplate("test_template.tf").Return(nil)
				return loader
			},
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:    "load single file when occurred loading error",
			Command: "./tflint test_template.tf",
			LoaderGenerator: func(ctrl *gomock.Controller) loader.LoaderIF {
				loader := mock.NewMockLoaderIF(ctrl)
				loader.EXPECT().LoadState()
				loader.EXPECT().LoadTFVars([]string{"terraform.tfvars"})
				loader.EXPECT().LoadTemplate("test_template.tf").Return(errors.New("loading error!"))
				return loader
			},
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status:     ExitCodeError,
				Stderr:     "loading error!",
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:              "enable debug option",
			Command:           "./tflint --debug",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          true,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
						Varfile:        []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "specify format",
			Command:           "./tflint --format json",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator: func(ctrl *gomock.Controller) printer.PrinterIF {
				printer := mock.NewMockPrinterIF(ctrl)
				printer.EXPECT().Print([]*issue.Issue{}, "json")
				return printer
			},
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:              "specify invalid format",
			Command:           "./tflint --format awesome",
			LoaderGenerator:   func(ctrl *gomock.Controller) loader.LoaderIF { return mock.NewMockLoaderIF(ctrl) },
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status: ExitCodeError,
				Stderr: "Invalid value `awesome' for option `-f, --format'",
			},
		},
		{
			Name:              "ignore_rules",
			Command:           "./tflint --ignore-rule rule1,rule2",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{"rule1": true, "rule2": true},
						Varfile:        []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "ignore_modules",
			Command:           "./tflint --ignore-module module1,module2",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{"module1": true, "module2": true},
						IgnoreRule:     map[string]bool{},
						Varfile:        []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:    "variable file",
			Command: "./tflint --var-file example1.tfvars,example2.tfvars",
			LoaderGenerator: func(ctrl *gomock.Controller) loader.LoaderIF {
				loader := mock.NewMockLoaderIF(ctrl)
				loader.EXPECT().LoadState()
				loader.EXPECT().LoadTFVars([]string{"terraform.tfvars", "example1.tfvars", "example2.tfvars"})
				loader.EXPECT().LoadAllTemplate(".").Return(nil)
				return loader
			},
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
						Varfile:        []string{"terraform.tfvars", "example1.tfvars", "example2.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "specify config gile",
			Command:           "./tflint --config .tflint.example.hcl",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
						Varfile:        []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.example.hcl",
				},
			},
		},
		{
			Name:              "enable deep check mode",
			Command:           "./tflint --deep",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      true,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
						Varfile:        []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "set aws static credentials",
			Command:           "./tflint --deep --aws-access-key AWS_ACCESS_KEY_ID --aws-secret-key AWS_SECRET_ACCESS_KEY --aws-region us-east-1",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
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
						Varfile:      []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "set aws shared credentials",
			Command:           "./tflint --deep --aws-profile account1 --aws-region us-east-1",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:     false,
						DeepCheck: true,
						AwsCredentials: map[string]string{
							"profile": "account1",
							"region":  "us-east-1",
						},
						IgnoreModule: map[string]bool{},
						IgnoreRule:   map[string]bool{},
						Varfile:      []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "enabled error with issues flag when no issues found",
			Command:           "./tflint --error-with-issues",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:            "enabled error with issues flag when issues found",
			Command:         "./tflint --error-with-issues",
			LoaderGenerator: loaderDefaultBehavior,
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF {
				detector := mock.NewMockDetectorIF(ctrl)
				detector.EXPECT().Detect().Return([]*issue.Issue{
					{
						Type:    "TEST",
						Message: "this is test method",
						Line:    1,
						File:    "",
					},
				})
				detector.EXPECT().HasError().Return(false)
				return detector
			},
			PrinterGenerator: func(ctrl *gomock.Controller) printer.PrinterIF {
				printer := mock.NewMockPrinterIF(ctrl)
				printer.EXPECT().Print([]*issue.Issue{
					{
						Type:    "TEST",
						Message: "this is test method",
						Line:    1,
						File:    "",
					},
				}, "default")
				return printer
			},
			Result: Result{
				Status:     ExitCodeIssuesFound,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:              "enable fast mode",
			Command:           "./tflint --fast",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerNoIssuesDefaultBehaviour,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{"aws_instance_invalid_ami": true},
						Varfile:        []string{"terraform.tfvars"},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "invalid options",
			Command:           "./tflint --debug --invalid-option",
			LoaderGenerator:   func(ctrl *gomock.Controller) loader.LoaderIF { return mock.NewMockLoaderIF(ctrl) },
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status: ExitCodeError,
				Stderr: "`invalid-option` is unknown option",
			},
		},
	}

	for _, tc := range cases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{
			outStream: outStream,
			errStream: errStream,
			loader:    tc.LoaderGenerator(ctrl),
			detector:  tc.DetectorGenerator(ctrl),
			printer:   tc.PrinterGenerator(ctrl),
			testMode:  true,
		}
		args := strings.Split(tc.Command, " ")

		status := cli.Run(args)
		if status != tc.Result.Status {
			t.Fatalf("˜\nBad: %d\nExpected: %d\n\ntestcase: %s", status, tc.Result.Status, tc.Name)
		}
		if !strings.Contains(outStream.String(), tc.Result.Stdout) {
			t.Fatalf("\nBad: %s\nExpected Contains: %s\n\ntestcase: %s", outStream.String(), tc.Result.Stdout, tc.Name)
		}
		if !strings.Contains(errStream.String(), tc.Result.Stderr) {
			t.Fatalf("\nBad: %s\nExpected Contains: %s\n\ntestcase: %s", errStream.String(), tc.Result.Stderr, tc.Name)
		}
		if !reflect.DeepEqual(cli.TestCLIOptions, tc.Result.CLIOptions) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(cli.TestCLIOptions), pp.Sprint(tc.Result.CLIOptions), tc.Name)
		}
	}
}
