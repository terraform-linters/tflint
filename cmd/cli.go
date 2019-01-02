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
	Cfg                  *tflint.Config
	opts                 Options
	argFiles             []string
	ExitCode             int
	noOp                 bool

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

func (cli *CLI) Run() int {
	if cli.noOp == true {
		return cli.ExitCode
	}

	cli.ProcessRules()
	cli.ReportViolations()

	return cli.ExitCode
}

func (cli *CLI) SanityCheck(args []string) error {
	parser := flags.NewParser(&cli.opts, flags.HelpFlag)
	parser.Usage = "[OPTIONS]"
	parser.UnknownOptionHandler = func(option string, arg flags.SplitArgument, args []string) ([]string, error) {
		if option == "debug" {
			cli.noOp = true
			return []string{}, errors.New("`debug` option was removed in v0.8.0. Please set `TFLINT_LOG` environment variables instead")
		}

		cli.noOp = true
		return []string{}, fmt.Errorf("`%s` is unknown option. Please run `tflint --help`", option)
	}

	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			cli.ExitCode = ExitCodeOK
			cli.noOp = true
			return nil
		}

		cli.printError(err)
		cli.ExitCode = ExitCodeError
		return err
	}

	cli.argFiles = args[1:]

	// Show version
	if cli.opts.Version {
		fmt.Fprintf(cli.outStream, "TFLint version %s\n", Version)
		cli.ExitCode = ExitCodeOK
		cli.noOp = true
		return nil
	}

	// Setup config
	cfg, err := tflint.LoadConfig(cli.opts.Config)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load TFLint config: %s", err))
		cli.ExitCode = ExitCodeError
		return err
	}

	cfg = cfg.Merge(cli.opts.toConfig())
	cli.Cfg = cfg

	// Load Terraform's configurations
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader()
		if err != nil {
			cli.printError(fmt.Errorf("Failed to prepare loading: %s", err))
			cli.ExitCode = ExitCodeError
			return err
		}
	}

	// Check to see if the all the files are correct
	for _, file := range cli.argFiles {
		if fileInfo, err := os.Stat(file); os.IsNotExist(err) {
			cli.printError(fmt.Errorf("Failed to load `%s`: File not found", file))
			cli.ExitCode = ExitCodeError
			return err
		} else if fileInfo.IsDir() {
			err = fmt.Errorf("Failed to load `%s`: TFLint doesn't accept directories as arguments", file)
			cli.printError(err)
			cli.ExitCode = ExitCodeError
			return err
		}

		if !cli.loader.IsConfigFile(file) {
			err = fmt.Errorf("Failed to load `%s`: File is not a target of Terraform", file)
			cli.printError(err)
			cli.ExitCode = ExitCodeError
			return err
		}
	}

	configs, err := cli.loader.LoadConfig()
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load configurations: %s", err))
		cli.ExitCode = ExitCodeError
		return err
	}

	valuesFiles, err := cli.loader.LoadValuesFiles(cfg.Varfile...)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load values files: %s", err))
		cli.ExitCode = ExitCodeError
		return err
	}

	// Check configurations via Runner
	runner := tflint.NewRunner(cfg, configs, valuesFiles...)
	cli.runners, err = tflint.NewModuleRunners(runner)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to prepare rule checking: %s", err))
		cli.ExitCode = ExitCodeError
		return err
	}

	cli.runners = append(cli.runners, runner)
	cli.ExitCode = ExitCodeOK
	return nil
}

func (cli *CLI) ProcessRules(additionalRules ...[]rules.Rule) {
	for _, rule := range rules.NewRules(cli.Cfg) {
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

func (cli *CLI) ReportViolations() {
	if cli.ExitCode == ExitCodeError {
		return
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
