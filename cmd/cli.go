package cmd

import (
	"errors"
	"fmt"
	"io"

	flags "github.com/jessevdk/go-flags"

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
}

// NewCLI returns new CLI initialized by input streams
func NewCLI(outStream io.Writer, errStream io.Writer) *CLI {
	return &CLI{
		outStream: outStream,
		errStream: errStream,
	}
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var opts Options
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
		fmt.Fprintf(cli.outStream, "TFLint version %s\n", Version)
		return ExitCodeOK
	}

	// Setup config
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	cfg = cfg.Merge(opts.toConfig())

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
	runner := tflint.NewRunner(cfg, configs)
	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	runners = append(runners, runner)

	for _, rule := range rules.NewRules(cfg) {
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
