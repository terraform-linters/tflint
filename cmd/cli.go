package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	flags "github.com/jessevdk/go-flags"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/printer"
	"github.com/wata727/tflint/project"
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
		if option == "debug" {
			return []string{}, errors.New("`debug` option was removed in v0.8.0. Please set `TFLINT_LOG` environment variables instead")
		}
		return []string{}, fmt.Errorf("`%s` is unknown option. Please run `tflint --help`", option)
	}
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			return ExitCodeOK
		}
		cli.printError(err)
		return ExitCodeError
	}
	argFiles := args[1:]

	// Show version
	if opts.Version {
		fmt.Fprintf(cli.outStream, "TFLint version %s\n", project.Version)
		return ExitCodeOK
	}

	// Setup config
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load TFLint config: %s", err))
		return ExitCodeError
	}
	cfg = cfg.Merge(opts.toConfig())

	// Load Terraform's configurations
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader()
		if err != nil {
			cli.printError(fmt.Errorf("Failed to prepare loading: %s", err))
			return ExitCodeError
		}
	}
	for _, file := range argFiles {
		if fileInfo, err := os.Stat(file); os.IsNotExist(err) {
			cli.printError(fmt.Errorf("Failed to load `%s`: File not found", file))
			return ExitCodeError
		} else if fileInfo.IsDir() {
			cli.printError(fmt.Errorf("Failed to load `%s`: TFLint doesn't accept directories as arguments", file))
			return ExitCodeError
		}

		if !cli.loader.IsConfigFile(file) {
			cli.printError(fmt.Errorf("Failed to load `%s`: File is not a target of Terraform", file))
			return ExitCodeError
		}
	}
	configs, err := cli.loader.LoadConfig()
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load configurations: %s", err))
		return ExitCodeError
	}
	valuesFiles, err := cli.loader.LoadValuesFiles(cfg.Varfile...)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load values files: %s", err))
		return ExitCodeError
	}

	// Check configurations via Runner
	runner := tflint.NewRunner(cfg, configs, valuesFiles...)
	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to prepare rule checking: %s", err))
		return ExitCodeError
	}
	runners = append(runners, runner)

	for _, rule := range rules.NewRules(cfg) {
		for _, runner := range runners {
			err := rule.Check(runner)
			if err != nil {
				cli.printError(fmt.Errorf("Failed to check `%s` rule: %s", rule.Name(), err))
				return ExitCodeError
			}
		}
	}

	issues := []*issue.Issue{}
	for _, runner := range runners {
		issues = append(issues, runner.LookupIssues(argFiles...)...)
	}

	// Print issues
	printer.NewPrinter(cli.outStream, cli.errStream).Print(issues, opts.Format, opts.Quiet)

	if opts.ErrorWithIssues && len(issues) > 0 {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) printError(err error) {
	fmt.Fprintln(cli.errStream, color.New(color.FgRed).Sprintf("Error: ")+err.Error())
}
