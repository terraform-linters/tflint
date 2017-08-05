package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"

	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/detector"
	"github.com/wata727/tflint/loader"
	"github.com/wata727/tflint/printer"
	"github.com/wata727/tflint/schema"
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
	Version         bool   `short:"v" long:"version" description:"Print TFLint version"`
	Format          string `short:"f" long:"format" description:"Output format" choice:"default" choice:"json" choice:"checkstyle" default:"default"`
	Config          string `short:"c" long:"config" description:"Config file name" value-name:"FILE" default:".tflint.hcl"`
	Recurse         bool   `short:"r" long:"recursive" description:"Descend recursively into subdirectories starting from cwd"`
	IgnoreModule    string `long:"ignore-module" description:"Ignore module sources" value-name:"SOURCE1,SOURCE2..."`
	IgnoreRule      string `long:"ignore-rule" description:"Ignore rule names" value-name:"RULE1,RULE2..."`
	Varfile         string `long:"var-file" description:"Terraform variable file names" value-name:"FILE1,FILE2..."`
	Deep            bool   `long:"deep" description:"Enable deep check mode"`
	AwsAccessKey    string `long:"aws-access-key" description:"AWS access key used in deep check mode" value-name:"ACCESS_KEY"`
	AwsSecretKey    string `long:"aws-secret-key" description:"AWS secret key used in deep check mode" value-name:"SECRET_KEY"`
	AwsProfile      string `long:"aws-profile" description:"AWS shared credential profile name used in deep check mode" value-name:"PROFILE"`
	AwsRegion       string `long:"aws-region" description:"AWS region used in deep check mode" value-name:"REGION"`
	Debug           bool   `short:"d" long:"debug" description:"Enable debug mode"`
	ErrorWithIssues bool   `long:"error-with-issues" description:"Return error code when issues exist"`
	Fast            bool   `long:"fast" description:"Ignore slow rules. Currently, ignore only aws_instance_invalid_ami"`
}

type TestCLIOptions struct {
	Config     *config.Config
	ConfigFile string
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var opts CLIOptions
	parser := flags.NewParser(&opts, flags.HelpFlag)
	parser.Usage = "[OPTIONS] [FILE]"
	parser.UnknownOptionHandler = func(option string, arg flags.SplitArgument, args []string) ([]string, error) {
		return []string{}, errors.New(fmt.Sprintf("ERROR: `%s` is unknown option. Please run `tflint --help`", option))
	}
	// Parse commandline flag
	args, err := parser.ParseArgs(args)
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Fprintln(cli.outStream, err)
			return ExitCodeOK
		} else {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}
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

	root := "./"
	dirsToCheck := []string{root}
	rootInfo, err := os.Stat(root)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	if c.Recurse {
		err = filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() {
				if f.Name() == ".git" || f.Name() == ".terraform" {
					return filepath.SkipDir
				}
				if !os.SameFile(f, rootInfo) {
					dirsToCheck = append(dirsToCheck, path)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}
	}

	for _, dir := range dirsToCheck {
		if c.Recurse {
			fmt.Printf("Checking dir: %s\n", dir)
		}

		owd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}
		err = os.Chdir(dir)
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
			templates, files, state, tfvars := cli.loader.Dump()
			schema, err := schema.Make(files)
			if err != nil {
				fmt.Fprintln(cli.errStream, fmt.Errorf("ERROR: Parse error: %s", err))
				return ExitCodeError
			}
			cli.detector, err = detector.NewDetector(templates, schema, state, tfvars, c)
			if err != nil {
				fmt.Fprintln(cli.errStream, err)
				return ExitCodeError
			}
		}

		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}

		issues := cli.detector.Detect()
		if cli.detector.HasError() {
			fmt.Fprintln(cli.errStream, "ERROR: error occurred in detecting. Please run with --debug option for details.")
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

		err = os.Chdir(owd)
		if err != nil {
			fmt.Fprintln(cli.errStream, err)
			return ExitCodeError
		}
	}

	return ExitCodeOK
}

func (cli *CLI) setupConfig(opts CLIOptions) (*config.Config, error) {
	c := config.Init()
	if opts.Debug {
		c.Debug = true
	}
	if opts.Recurse {
		c.Recurse = true
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
