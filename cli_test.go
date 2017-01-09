package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
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
		loader.EXPECT().LoadAllFile(".").Return(nil)
		return loader
	}
	var detectorNoErrorNoIssuesBehavior = func(ctrl *gomock.Controller) detector.DetectorIF {
		detector := mock.NewMockDetectorIF(ctrl)
		detector.EXPECT().Detect().Return([]*issue.Issue{})
		detector.EXPECT().HasError().Return(false)
		return detector
	}
	var printerDefaultBehavior = func(ctrl *gomock.Controller) printer.PrinterIF {
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
			Name:              "print version by alias",
			Command:           "./tflint -v",
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
				Stdout: "Usage: tflint [<options>] <args>",
			},
		},
		{
			Name:              "print help by alias",
			Command:           "./tflint --h",
			LoaderGenerator:   func(ctrl *gomock.Controller) loader.LoaderIF { return mock.NewMockLoaderIF(ctrl) },
			DetectorGenerator: func(ctrl *gomock.Controller) detector.DetectorIF { return mock.NewMockDetectorIF(ctrl) },
			PrinterGenerator:  func(ctrl *gomock.Controller) printer.PrinterIF { return mock.NewMockPrinterIF(ctrl) },
			Result: Result{
				Status: ExitCodeOK,
				Stdout: "Usage: tflint [<options>] <args>",
			},
		},
		{
			Name:              "nothing options",
			Command:           "./tflint",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:    "load single file",
			Command: "./tflint test_template.tf",
			LoaderGenerator: func(ctrl *gomock.Controller) loader.LoaderIF {
				loader := mock.NewMockLoaderIF(ctrl)
				loader.EXPECT().LoadFile("test_template.tf").Return(nil)
				return loader
			},
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status:     ExitCodeOK,
				CLIOptions: defaultCLIOptions,
			},
		},
		{
			Name:              "enable debug option",
			Command:           "./tflint --debug",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          true,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "enable debug option by alias",
			Command:           "./tflint -d",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          true,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "specify json format",
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
			Name:              "specify json format by alias",
			Command:           "./tflint -f json",
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
			Name:              "ignore_rules",
			Command:           "./tflint --ignore-rule rule1,rule2",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{"rule1": true, "rule2": true},
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
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{"module1": true, "module2": true},
						IgnoreRule:     map[string]bool{},
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
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					ConfigFile: ".tflint.example.hcl",
				},
			},
		},
		{
			Name:              "specify config gile by alias",
			Command:           "./tflint -c .tflint.example.hcl",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      false,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
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
			PrinterGenerator:  printerDefaultBehavior,
			Result: Result{
				Status: ExitCodeOK,
				CLIOptions: TestCLIOptions{
					Config: &config.Config{
						Debug:          false,
						DeepCheck:      true,
						AwsCredentials: map[string]string{},
						IgnoreModule:   map[string]bool{},
						IgnoreRule:     map[string]bool{},
					},
					ConfigFile: ".tflint.hcl",
				},
			},
		},
		{
			Name:              "set aws credentials",
			Command:           "./tflint --deep --aws-access-key AWS_ACCESS_KEY_ID --aws-secret-key AWS_SECRET_ACCESS_KEY --aws-region us-east-1",
			LoaderGenerator:   loaderDefaultBehavior,
			DetectorGenerator: detectorNoErrorNoIssuesBehavior,
			PrinterGenerator:  printerDefaultBehavior,
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
					},
					ConfigFile: ".tflint.hcl",
				},
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
			t.Fatalf("Bad: %d\nExpected: %d\n\ntestcase: %s", status, tc.Result.Status, tc.Name)
		}
		if !strings.Contains(outStream.String(), tc.Result.Stdout) {
			t.Fatalf("Bad: %s\nExpected Contains: %s\n\ntestcase: %s", outStream.String(), tc.Result.Stdout, tc.Name)
		}
		if !strings.Contains(errStream.String(), tc.Result.Stderr) {
			t.Fatalf("Bad: %s\nExpected Contains: %s\n\ntestcase: %s", errStream.String(), tc.Result.Stderr, tc.Name)
		}
		if !reflect.DeepEqual(cli.TestCLIOptions, tc.Result.CLIOptions) {
			t.Fatalf("Bad: %+v\nExpected: %+v\n\ntestcase: %s", cli.TestCLIOptions, tc.Result.CLIOptions, tc.Name)
		}
	}
}
