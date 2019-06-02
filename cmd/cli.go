package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	parser.Usage = "[OPTIONS] [FILE or DIR...]"
	parser.UnknownOptionHandler = func(option string, arg flags.SplitArgument, args []string) ([]string, error) {
		if option == "debug" {
			return []string{}, errors.New("`debug` option was removed in v0.8.0. Please set `TFLINT_LOG` environment variables instead")
		}
		if option == "fast" {
			return []string{}, errors.New("`fast` option was removed in v0.9.0. The `aws_instance_invalid_ami` rule is already fast enough")
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
	dir, filterFiles, err := processArgs(args[1:])
	if err != nil {
		cli.printError(err)
		return ExitCodeError
	}

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

	configs, err := cli.loader.LoadConfig(dir)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load configurations: %s", err))
		return ExitCodeError
	}
	annotations, err := cli.loader.LoadAnnotations(dir)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load configuration tokens: %s", err))
		return ExitCodeError
	}
	variables, err := cli.loader.LoadValuesFiles(cfg.Varfile...)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to load values files: %s", err))
		return ExitCodeError
	}
	cliVars, err := tflint.ParseTFVariables(cfg.Variables, configs.Module.Variables)
	if err != nil {
		cli.printError(fmt.Errorf("Failed to parse variables: %s", err))
		return ExitCodeError
	}
	variables = append(variables, cliVars)

	// Check configurations via Runner
	runner := tflint.NewRunner(cfg, annotations, configs, variables...)
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
		issues = append(issues, runner.LookupIssues(filterFiles...)...)
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

func processArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return ".", []string{}, nil
	}

	var dir string
	filterFiles := []string{}

	for _, file := range args {
		fileInfo, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				return dir, filterFiles, fmt.Errorf("Failed to load `%s`: File not found", file)
			}
			return dir, filterFiles, fmt.Errorf("Failed to load `%s`: %s", file, err)
		}

		if fileInfo.IsDir() {
			dir = file
			if len(args) != 1 {
				return dir, filterFiles, fmt.Errorf("Failed to load `%s`: Multiple arguments are not allowed when passing a directory", file)
			}
			return dir, filterFiles, nil
		}

		if !strings.HasSuffix(file, ".tf") && !strings.HasSuffix(file, ".tf.json") {
			return dir, filterFiles, fmt.Errorf("Failed to load `%s`: File is not a target of Terraform", file)
		}

		fileDir := filepath.Dir(file)
		if dir == "" {
			dir = fileDir
			filterFiles = append(filterFiles, file)
		} else if fileDir == dir {
			filterFiles = append(filterFiles, file)
		} else {
			return dir, filterFiles, fmt.Errorf("Failed to load `%s`: Multiple files in different directories are not allowed", file)
		}
	}

	return dir, filterFiles, nil
}
