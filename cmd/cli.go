package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	flags "github.com/jessevdk/go-flags"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/afero"

	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/langserver"
	"github.com/terraform-linters/tflint/rules"
	"github.com/terraform-linters/tflint/tflint"
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
		if option == "error-with-issues" {
			return []string{}, errors.New("`error-with-issues` option was removed in v0.9.0. The behavior is now default")
		}
		if option == "quiet" || option == "q" {
			return []string{}, errors.New("`quiet` option was removed in v0.11.0. The behavior is now default")
		}
		if option == "ignore-rule" {
			return []string{}, errors.New("`ignore-rule` option was removed in v0.12.0. Please use `--disable-rule` instead")
		}
		return []string{}, fmt.Errorf("`%s` is unknown option. Please run `tflint --help`", option)
	}
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	// Set up output formatter
	formatter := &formatter.Formatter{
		Stdout: cli.outStream,
		Stderr: cli.errStream,
		Format: opts.Format,
	}
	if opts.NoColor {
		color.NoColor = true
		formatter.NoColor = true
	}

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			return ExitCodeOK
		}
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to parse CLI options", err), map[string][]byte{})
		return ExitCodeError
	}
	dir, filterFiles, err := processArgs(args[1:])
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to parse CLI arguments", err), map[string][]byte{})
		return ExitCodeError
	}

	// Show version
	if opts.Version {
		fmt.Fprintf(cli.outStream, "TFLint version %s\n", tflint.Version)
		return ExitCodeOK
	}

	// Start language server
	if opts.Langserver {
		return cli.startServer(opts.Config, opts.toConfig())
	}

	// Setup config
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to load TFLint config", err), map[string][]byte{})
		return ExitCodeError
	}
	cfg = cfg.Merge(opts.toConfig())

	// Load Terraform's configurations
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, cfg)
		if err != nil {
			formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to prepare loading", err), map[string][]byte{})
			return ExitCodeError
		}
	}

	configs, err := cli.loader.LoadConfig(dir)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to load configurations", err), cli.loader.Sources())
		return ExitCodeError
	}
	annotations, err := cli.loader.LoadAnnotations(dir)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to load configuration tokens", err), cli.loader.Sources())
		return ExitCodeError
	}
	variables, err := cli.loader.LoadValuesFiles(cfg.Varfiles...)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to load values files", err), cli.loader.Sources())
		return ExitCodeError
	}
	cliVars, err := tflint.ParseTFVariables(cfg.Variables, configs.Module.Variables)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to parse variables", err), cli.loader.Sources())
		return ExitCodeError
	}
	variables = append(variables, cliVars)

	// Check configurations via Runner
	runner, err := tflint.NewRunner(cfg, annotations, configs, variables...)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to initialize a runner", err), cli.loader.Sources())
		return ExitCodeError
	}
	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to prepare rule checking", err), cli.loader.Sources())
		return ExitCodeError
	}
	runners = append(runners, runner)

	var ruleNames []string
	for name := range cfg.Rules {
		ruleNames = append(ruleNames, name)
	}
	err = rules.CheckRuleNames(ruleNames)
	if err != nil {
		formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to check rule config", err), cli.loader.Sources())
		return ExitCodeError
	}

	for _, rule := range rules.NewRules(cfg) {
		for _, runner := range runners {
			err := rule.Check(runner)
			if err != nil {
				formatter.Print(tflint.Issues{}, tflint.NewContextError(fmt.Sprintf("Failed to check `%s` rule", rule.Name()), err), cli.loader.Sources())
				return ExitCodeError
			}
		}
	}

	issues := tflint.Issues{}
	for _, runner := range runners {
		issues = append(issues, runner.LookupIssues(filterFiles...)...)
	}

	// Print issues
	formatter.Print(issues, nil, cli.loader.Sources())

	if len(issues) > 0 && !cfg.Force {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) startServer(configPath string, cliConfig *tflint.Config) int {
	log.Println("Starting language server...")

	handler, err := langserver.NewHandler(configPath, cliConfig)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to start language server: %s", err))
		return ExitCodeError
	}

	var connOpt []jsonrpc2.ConnOpt
	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(langserver.NewConn(os.Stdin, os.Stdout), jsonrpc2.VSCodeObjectCodec{}),
		handler,
		connOpt...,
	).DisconnectNotify()
	log.Println("Shutting down...")

	return ExitCodeOK
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
