package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
	flags "github.com/jessevdk/go-flags"
	"github.com/terraform-linters/tflint/formatter"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK int = iota
	ExitCodeError
	ExitCodeIssuesFound
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
	originalWorkingDir   string
	sources              map[string][]byte

	// fields for each module
	config    *tflint.Config
	loader    *terraform.Loader
	formatter *formatter.Formatter
}

// NewCLI returns new CLI initialized by input streams
func NewCLI(outStream io.Writer, errStream io.Writer) (*CLI, error) {
	wd, err := os.Getwd()

	return &CLI{
		outStream:          outStream,
		errStream:          errStream,
		originalWorkingDir: wd,
		sources:            map[string][]byte{},
	}, err
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var opts Options
	parser := flags.NewParser(&opts, flags.HelpFlag)
	parser.Usage = "--chdir=DIR/--recursive [OPTIONS]"
	parser.UnknownOptionHandler = unknownOptionHandler
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	// Set up output formatter
	cli.formatter = &formatter.Formatter{
		Stdout: cli.outStream,
		Stderr: cli.errStream,
		Format: opts.Format,
	}
	if opts.Color {
		color.NoColor = false
		cli.formatter.NoColor = false
	}
	if opts.NoColor {
		color.NoColor = true
		cli.formatter.NoColor = true
	}
	level := os.Getenv("TFLINT_LOG")
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(level)),
		Writer:   os.Stderr,
	})
	log.SetFlags(log.Ltime | log.Lshortfile)

	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			return ExitCodeOK
		}
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to parse CLI options; %w", err), map[string][]byte{})
		return ExitCodeError
	}
	if len(args) > 1 {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Command line arguments support was dropped in v0.47. Use --chdir or --filter instead."), map[string][]byte{})
		return ExitCodeError
	}

	switch {
	case opts.Version:
		return cli.printVersion(opts)
	case opts.Init:
		return cli.init(opts)
	case opts.Langserver:
		return cli.startLanguageServer(opts)
	case opts.ActAsBundledPlugin:
		return cli.actAsBundledPlugin()
	default:
		return cli.inspect(opts)
	}
}

func unknownOptionHandler(option string, arg flags.SplitArgument, args []string) ([]string, error) {
	if option == "debug" {
		return []string{}, errors.New("--debug option was removed in v0.8.0. Please set TFLINT_LOG environment variables instead")
	}
	if option == "error-with-issues" {
		return []string{}, errors.New("--error-with-issues option was removed in v0.9.0. The behavior is now default")
	}
	if option == "quiet" || option == "q" {
		return []string{}, errors.New("--quiet option was removed in v0.11.0. The behavior is now default")
	}
	if option == "ignore-rule" {
		return []string{}, errors.New("--ignore-rule option was removed in v0.12.0. Please use --disable-rule instead")
	}
	if option == "deep" {
		return []string{}, errors.New("--deep option was removed in v0.23.0. Deep checking is now a feature of the AWS plugin, so please configure the plugin instead")
	}
	if option == "aws-access-key" || option == "aws-secret-key" || option == "aws-profile" || option == "aws-creds-file" || option == "aws-region" {
		return []string{}, fmt.Errorf("--%s option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead", option)
	}
	if option == "loglevel" {
		return []string{}, errors.New("--loglevel option was removed in v0.40.0. Please set TFLINT_LOG environment variables instead")
	}
	return []string{}, fmt.Errorf(`--%s is unknown option. Please run "tflint --help"`, option)
}

func findWorkingDirs(opts Options) ([]string, error) {
	if opts.Recursive && opts.Chdir != "" {
		return []string{}, errors.New("cannot use --recursive and --chdir at the same time")
	}

	workingDirs := []string{}

	if opts.Recursive {
		// NOTE: The target directory is always the current directory in recursive mode
		err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}
			// hidden directories are skipped
			if path != "." && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			workingDirs = append(workingDirs, path)
			return nil
		})
		if err != nil {
			return []string{}, err
		}
	} else {
		if opts.Chdir == "" {
			workingDirs = []string{"."}
		} else {
			workingDirs = []string{opts.Chdir}
		}
	}

	return workingDirs, nil
}

func (cli *CLI) withinChangedDir(dir string, proc func() error) (err error) {
	if dir != "." {
		chErr := os.Chdir(dir)
		if chErr != nil {
			return fmt.Errorf("Failed to switch to a different working directory; %w", chErr)
		}
		defer func() {
			chErr := os.Chdir(cli.originalWorkingDir)
			if chErr != nil {
				err = fmt.Errorf("Failed to switch to the original working directory; %s; %w", chErr, err)
			}
		}()
	}

	return proc()
}
