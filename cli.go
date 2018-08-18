package main

import (
	"errors"
	"fmt"
	"io"

	flags "github.com/jessevdk/go-flags"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/printer"
	"github.com/wata727/tflint/rules"
	"github.com/wata727/tflint/tflint"
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
	loader               tflint.AbstractLoader
	testMode             bool
	TestCLIOptions       TestCLIOptions
}

// CLIOptions is an option specified by arguments.
type CLIOptions struct {
	Version         bool   `short:"v" long:"version" description:"Print TFLint version"`
	Format          string `short:"f" long:"format" description:"Output format" choice:"default" choice:"json" choice:"checkstyle" default:"default"`
	Config          string `short:"c" long:"config" description:"Config file name" value-name:"FILE" default:".tflint.hcl"`
	IgnoreModule    string `long:"ignore-module" description:"Ignore module sources" value-name:"SOURCE1,SOURCE2..."`
	IgnoreRule      string `long:"ignore-rule" description:"Ignore rule names" value-name:"RULE1,RULE2..."`
	Varfile         string `long:"var-file" description:"Terraform variable file names" value-name:"FILE1,FILE2..."`
	Deep            bool   `long:"deep" description:"Enable deep check mode"`
	AwsAccessKey    string `long:"aws-access-key" description:"AWS access key used in deep check mode" value-name:"ACCESS_KEY"`
	AwsSecretKey    string `long:"aws-secret-key" description:"AWS secret key used in deep check mode" value-name:"SECRET_KEY"`
	AwsProfile      string `long:"aws-profile" description:"AWS shared credential profile name used in deep check mode" value-name:"PROFILE"`
	AwsRegion       string `long:"aws-region" description:"AWS region used in deep check mode" value-name:"REGION"`
	ErrorWithIssues bool   `long:"error-with-issues" description:"Return error code when issues exist"`
	Fast            bool   `long:"fast" description:"Ignore slow rules (aws_instance_invalid_ami only)"`
	Quiet           bool   `short:"q" long:"quiet" description:"Do not output any message when no issues are found (default format only)"`
}

// TestCLIOptions is a set of configs assembled from the CLIOptions for test
type TestCLIOptions struct {
	Config     *config.Config
	ConfigFile string
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var opts CLIOptions
	parser := flags.NewParser(&opts, flags.HelpFlag)
	parser.Usage = "[OPTIONS]"
	parser.UnknownOptionHandler = func(option string, arg flags.SplitArgument, args []string) ([]string, error) {
		return []string{}, fmt.Errorf("ERROR: `%s` is unknown option. Please run `tflint --help`", option)
	}
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			return ExitCodeOK
		}
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	if len(args) > 1 {
		fmt.Fprintln(cli.errStream, errors.New("ERROR: Too many arguments. TFLint doesn't accept the file argument"))
		return ExitCodeError
	}

	// Show version
	if opts.Version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Setup config
	c, err := cli.setupConfig(opts)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}

	// Load Terraform's configurations
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader()
		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}
	}
	configs, err := cli.loader.LoadConfig()
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}

	// Check configurations via Runner
	runner := tflint.NewRunner(c, configs)
	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	runners = append(runners, runner)

	for _, rule := range rules.NewRules(c) {
		for _, runner := range runners {
			err := rule.Check(runner)
			if err != nil {
				fmt.Fprintln(cli.errStream, err)
				return ExitCodeError
			}
		}
	}

	issues := []*issue.Issue{}
	for _, runner := range runners {
		issues = append(issues, runner.Issues...)
	}

	// Print issues
	printer.NewPrinter(cli.outStream, cli.errStream).Print(issues, opts.Format, opts.Quiet)

	if opts.ErrorWithIssues && len(issues) > 0 {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) setupConfig(opts CLIOptions) (*config.Config, error) {
	c := config.Init()
	fallbackConfig, err := homedir.Expand("~/.tflint.hcl")
	if err != nil {
		return nil, err
	}
	if err := c.LoadConfig(opts.Config, fallbackConfig); err != nil {
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
