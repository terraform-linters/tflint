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
	cfg                  *tflint.Config
	opts                 Options
	argFiles             []string
	ExitCode             int

	runners []*tflint.Runner
	Issues  []*issue.Issue
}

// NewCLI returns new CLI initialized by input streams
func NewCLI(outStream io.Writer, errStream io.Writer) *CLI {
	return &CLI{
		outStream: outStream,
		errStream: errStream,
	}
}

func (cli *CLI) SanityCheck(args []string) {
	parser := flags.NewParser(&cli.opts, flags.HelpFlag)
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
			cli.ExitCode = ExitCodeOK
			return
		}
		cli.printError(err)
		cli.ExitCode = ExitCodeError
		return
	}

	cli.argFiles = args[1:]

	// Show version
	if cli.opts.Version {
		fmt.Fprintf(cli.outStream, "TFLint version %s\n", Version)
		cli.ExitCode = ExitCodeOK
		return
	}

	// Setup config
	cfg, err := tflint.LoadConfig(cli.opts.Config)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load TFLint config: %s", err))
		cli.ExitCode = ExitCodeError
		return
	}
	cfg = cfg.Merge(cli.opts.toConfig())
	cli.cfg = cfg

	// Load Terraform's configurations
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader()
		if err != nil {
			cli.printError(fmt.Errorf("Failed to prepare loading: %s", err))
			cli.ExitCode = ExitCodeError
			return
		}
	}
	// Check to see if the all the files are correct
	for _, file := range cli.argFiles {
		if fileInfo, err := os.Stat(file); os.IsNotExist(err) {
			cli.printError(fmt.Errorf("Failed to load `%s`: File not found", file))
			cli.ExitCode = ExitCodeError
			return
		} else if fileInfo.IsDir() {
			cli.printError(fmt.Errorf("Failed to load `%s`: TFLint doesn't accept directories as arguments", file))
			cli.ExitCode = ExitCodeError
			return
		}

		if !cli.loader.IsConfigFile(file) {
			cli.printError(fmt.Errorf("Failed to load `%s`: File is not a target of Terraform", file))
			cli.ExitCode = ExitCodeError
			return
		}
	}
	configs, err := cli.loader.LoadConfig()
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load configurations: %s", err))
		cli.ExitCode = ExitCodeError
		return
	}
	valuesFiles, err := cli.loader.LoadValuesFiles(cfg.Varfile...)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load values files: %s", err))
		cli.ExitCode = ExitCodeError
		return
	}

	// Check configurations via Runner
	runner := tflint.NewRunner(cfg, configs, valuesFiles...)
	cli.runners, err = tflint.NewModuleRunners(runner)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to prepare rule checking: %s", err))
		cli.ExitCode = ExitCodeError
		return
	}

	cli.runners = append(cli.runners, runner)
	cli.ExitCode = ExitCodeOK
	return
}

func (cli *CLI) ProcessRules(overrideRules *rules.OverrideRules) {
	for _, rule := range rules.NewRules(cli.cfg, overrideRules) {
		for _, runner := range cli.runners {
			err := rule.Check(runner)
			if err != nil {
				cli.printError(fmt.Errorf("Failed to check `%s` rule: %s", rule.Name(), err))
				cli.ExitCode = ExitCodeError
				return
			}
		}
	}

	cli.ExitCode = ExitCodeOK

	for _, runner := range cli.runners {
		cli.Issues = append(cli.Issues, runner.LookupIssues(cli.argFiles...)...)
	}
}

func (cli *CLI) ReportViolations(additionalIssues []*issue.Issue) {
	// Print issues
	for _, issue := range additionalIssues {
		cli.Issues = append(cli.Issues, issue)
	}

	printer.NewPrinter(cli.outStream, cli.errStream).Print(cli.Issues, cli.opts.Format, cli.opts.Quiet)

	if cli.opts.ErrorWithIssues && len(cli.Issues) > 0 {
		cli.ExitCode = ExitCodeIssuesFound
		return
	}

	cli.ExitCode = ExitCodeOK
	return
}

func (cli *CLI) printError(err error) {
	fmt.Fprintln(cli.errStream, color.New(color.FgRed).Sprintf("Error: ")+err.Error())
}
