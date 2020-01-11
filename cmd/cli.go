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

	"github.com/terraform-linters/tflint/formatter"
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
	formatter            *formatter.Formatter
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
	parser.UnknownOptionHandler = unknownOptionHandler
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	// Set up output formatter
	cli.formatter = &formatter.Formatter{
		Stdout: cli.outStream,
		Stderr: cli.errStream,
		Format: opts.Format,
	}
	if opts.NoColor {
		color.NoColor = true
		cli.formatter.NoColor = true
	}

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			return ExitCodeOK
		}
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to parse CLI options", err), map[string][]byte{})
		return ExitCodeError
	}
	dir, filterFiles, err := processArgs(args[1:])
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to parse CLI arguments", err), map[string][]byte{})
		return ExitCodeError
	}

	switch {
	case opts.Version:
		fmt.Fprintf(cli.outStream, "TFLint version %s\n", tflint.Version)
		return ExitCodeOK
	case opts.Langserver:
		return cli.startLanguageServer(opts.Config, opts.toConfig())
	default:
		return cli.inspect(opts, dir, filterFiles)
	}
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

func unknownOptionHandler(option string, arg flags.SplitArgument, args []string) ([]string, error) {
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
