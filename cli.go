package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/detector"
	"github.com/wata727/tflint/loader"
	"github.com/wata727/tflint/printer"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
	testMode             bool
	TestCLIOptions       TestCLIOptions
}

type TestCLIOptions struct {
	Config     *config.Config
	Format     string
	LoadFile   string
	ConfigFile string
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		version      bool
		help         bool
		debug        bool
		format       string
		ignoreModule string
		ignoreRule   string
		configFile   string
		deepCheck    bool
		awsAccessKey string
		awsSecretKey string
		awsRegion    string
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	// Do not print default usage message
	flags.SetOutput(new(bytes.Buffer))

	flags.BoolVar(&version, "version", false, "print version information.")
	flags.BoolVar(&version, "v", false, "alias for -version")
	flags.BoolVar(&help, "help", false, "show usage of TFLint. This page.")
	flags.BoolVar(&help, "h", false, "alias for --help")
	flags.BoolVar(&debug, "debug", false, "enable debug mode.")
	flags.BoolVar(&debug, "d", false, "alias for --debug")
	flags.StringVar(&format, "format", "default", "choose output format from \"default\" or \"json\"")
	flags.StringVar(&format, "f", "default", "alias for --format")
	flags.StringVar(&ignoreModule, "ignore-module", "", "ignore module by specified source.")
	flags.StringVar(&ignoreRule, "ignore-rule", "", "ignore rules.")
	flags.StringVar(&configFile, "config", ".tflint.hcl", "specify config file. default is \".tflint.hcl\"")
	flags.StringVar(&configFile, "c", ".tflint.hcl", "alias for --config")
	flags.BoolVar(&deepCheck, "deep", false, "enable deep check mode.")
	flags.StringVar(&awsAccessKey, "aws-access-key", "", "AWS access key used in deep check mode.")
	flags.StringVar(&awsSecretKey, "aws-secret-key", "", "AWS secret key used in deep check mode.")
	flags.StringVar(&awsRegion, "aws-region", "", "AWS region used in deep check mode.")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(cli.errStream, "ERROR: `%s` is unknown options. Please run `tflint --help`\n", args[1])
		return ExitCodeError
	}
	if !printer.ValidateFormat(format) {
		fmt.Fprintf(cli.errStream, "ERROR: `%s` is unknown format. Please run `tflint --help`\n", format)
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Show help
	if help {
		fmt.Fprintln(cli.outStream, `TFLint is a linter of Terraform.

Usage: tflint [<options>] <args>

Available options:
	-h, --help				show usage of TFLint. This page.
	-v, --version				print version information.
	-f, --format <format>			choose output format from "default" or "json"
	-c, --config <file>			specify config file. default is ".tflint.hcl"
	--ignore-module <source1,source2...>	ignore module by specified source.
	--ignore-rule <rule1,rule2...>		ignore rules.
	--deep					enable deep check mode.
	--aws-access-key			set AWS access key used in deep check mode.
	--aws-secret-key			set AWS secret key used in deep check mode.
	--aws-region				set AWS region used in deep check mode.
	-d, --debug				enable debug mode.

Support aruguments:
	TFLint scans all configuration file of Terraform in current directory by default.
	If you specified single file path, it scans only this.
`)
		return ExitCodeOK
	}

	// Setup config
	c := config.Init()
	if debug {
		c.Debug = true
	}
	if err := c.LoadConfig(configFile); err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	if deepCheck || c.DeepCheck {
		c.DeepCheck = true
		c.SetAwsCredentials(awsAccessKey, awsSecretKey, awsRegion)
	}
	if ignoreModule != "" {
		c.SetIgnoreModule(ignoreModule)
	}
	if ignoreRule != "" {
		c.SetIgnoreRule(ignoreRule)
	}

	if cli.testMode {
		cli.TestCLIOptions = TestCLIOptions{
			Config:     c,
			Format:     format,
			LoadFile:   flags.Arg(0),
			ConfigFile: configFile,
		}
	} else {
		// Main function
		var err error
		l := loader.NewLoader(c.Debug)
		if flags.NArg() > 0 {
			err = l.LoadFile(flags.Arg(0))
		} else {
			err = l.LoadAllFile(".")
		}
		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}

		d, err := detector.NewDetector(l.ListMap, c)
		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}
		issues := d.Detect()
		if d.Error {
			fmt.Fprintln(cli.errStream, "ERROR: error occurred in detecting. Please run with --debug options for details.")
			return ExitCodeError
		}

		p := printer.NewPrinter(cli.outStream, cli.errStream)
		p.Print(issues, format)
	}

	return ExitCodeOK
}
