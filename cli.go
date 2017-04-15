package main

import (
	"errors"
	"fmt"
	"io"

	flags "github.com/jessevdk/go-flags"

	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/detector"
	"github.com/wata727/tflint/loader"
	"github.com/wata727/tflint/printer"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
	ExitCodeIssuesFound
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
	loader               loader.LoaderIF
	detector             detector.DetectorIF
	printer              printer.PrinterIF
	testMode             bool
	TestCLIOptions       TestCLIOptions
}

type CLIOptions struct {
	Help            bool   `short:"h" long:"help" description:"Show usage."`
	Version         bool   `short:"v" long:"version" description:"Print TFLint version."`
	Format          string `short:"f" long:"format" description:"Choose the output of TFLint format from \"default\", \"json\" or \"checkstyle\"" default:"default"`
	Config          string `short:"c" long:"config" description:"Specify a config file name. default is \".tflint.hcl\"" default:".tflint.hcl"`
	IgnoreModule    string `long:"ignore-module" description:"Specify module names to be ignored, separated by commas."`
	IgnoreRule      string `long:"ignore-rule" description:"Specify rule names to be ignored, separated by commas."`
	Varfile         string `long:"var-file" description:"Specify Terraform variable file names, separated by commas."`
	Deep            bool   `long:"deep" description:"Enable deep check mode."`
	AwsAccessKey    string `long:"aws-access-key" description:"Set AWS access key used in deep check mode."`
	AwsSecretKey    string `long:"aws-secret-key" description:"Set AWS secret key used in deep check mode."`
	AwsProfile      string `long:"aws-profile" description:"Set AWS shared credential profile name used in deep check mode."`
	AwsRegion       string `long:"aws-region" description:"Set AWS region used in deep check mode."`
	Debug           bool   `short:"d" long:"debug" description:"Enable debug mode."`
	ErrorWithIssues bool   `long:"error-with-issues" description:"Return error code when issue exists."`
	Fast            bool   `long:"fast" description:"Ignore slow rules. Currently, ignore only aws_instance_invalid_ami"`
}

type TestCLIOptions struct {
	Config     *config.Config
	ConfigFile string
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var opts CLIOptions
	parser := flags.NewParser(&opts, flags.None)
	parser.UnknownOptionHandler = func(option string, arg flags.SplitArgument, args []string) ([]string, error) {
		return []string{}, errors.New(fmt.Sprintf("ERROR: `%s` is unknown option. Please run `tflint --help`\n", option))
	}
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	if err != nil {
		fmt.Fprint(cli.errStream, err)
		return ExitCodeError
	}

	if !printer.ValidateFormat(opts.Format) {
		fmt.Fprintf(cli.errStream, "ERROR: `%s` is unknown format. Please run `tflint --help`\n", opts.Format)
		return ExitCodeError
	}

	// Show version
	if opts.Version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Show help
	if opts.Help {
		fmt.Fprintln(cli.outStream, Help)
		return ExitCodeOK
	}

	// Setup config
	c, err := cli.setupConfig(opts)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}

	// Main function
	// If disabled test mode, generates real loader
	if !cli.testMode {
		cli.loader = loader.NewLoader(c.Debug)
	}
	cli.loader.LoadState()
	cli.loader.LoadTFVars(c.Varfile)
	if len(args) > 1 {
		err = cli.loader.LoadTemplate(args[1])
	} else {
		err = cli.loader.LoadAllTemplate(".")
	}
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}

	// If disabled test mode, generates real detector
	if !cli.testMode {
		templates, state, tfvars := cli.loader.Dump()
		cli.detector, err = detector.NewDetector(templates, state, tfvars, c)
	}
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	issues := cli.detector.Detect()
	if cli.detector.HasError() {
		fmt.Fprintln(cli.errStream, "ERROR: error occurred in detecting. Please run with --debug options for details.")
		return ExitCodeError
	}

	// If disabled test mode, generates real printer
	if !cli.testMode {
		cli.printer = printer.NewPrinter(cli.outStream, cli.errStream)
	}
	cli.printer.Print(issues, opts.Format)

	if opts.ErrorWithIssues && len(issues) > 0 {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) setupConfig(opts CLIOptions) (*config.Config, error) {
	c := config.Init()
	if opts.Debug {
		c.Debug = true
	}
	if err := c.LoadConfig(opts.Config); err != nil {
		return nil, err
	}
	if opts.Deep || c.DeepCheck {
		c.DeepCheck = true
		c.SetAwsCredentials(opts.AwsAccessKey, opts.AwsSecretKey, opts.AwsProfile, opts.AwsRegion)
	}
	// `aws_instance_invalid_ami` is very slow...
	if opts.Fast {
		c.SetIgnoreRule("aws_instance_invalid_ami")
	}
	c.SetIgnoreModule(opts.IgnoreModule)
	c.SetIgnoreRule(opts.IgnoreRule)
	c.SetVarfile(opts.Varfile)
	// If enabled test mode, set config information
	if cli.testMode {
		cli.TestCLIOptions = TestCLIOptions{
			Config:     c,
			ConfigFile: opts.Config,
		}
	}
	return c, nil
}
