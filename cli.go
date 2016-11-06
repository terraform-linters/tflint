package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/detector"
	"github.com/wata727/tflint/loader"
	"github.com/wata727/tflint/printer"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		version bool
		help    bool
		debug   bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	// Do not print default usage message
	flags.SetOutput(new(bytes.Buffer))

	flags.BoolVar(&version, "version", false, "Print version information and quit.")
	flags.BoolVar(&version, "v", false, "Alias for -version")
	flags.BoolVar(&help, "help", false, "Show usage (this page)")
	flags.BoolVar(&help, "h", false, "Alias for -help")
	flags.BoolVar(&debug, "debug", false, "Enable debug mode")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(cli.errStream, "ERROR: `%s` is unknown options. Please run `tflint --help`\n", args[1])
		return ExitCodeError
	}

	c := config.Init()

	// Show version
	if version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	if help {
		fmt.Fprintln(cli.outStream, `TFLint is a linter of Terraform.

Usage: tflint [<options>] <args>

Available options:
	-h, --help	show usage of TFLint. This page.
	-v, --version	print version information.
	--debug		enable debug mode.

Support aruguments:
	TFLint scans all configuration file of Terraform in current directory by default.
	If you specified single file path, it scans only this.
`)
		return ExitCodeOK
	}

	if debug {
		c.Debug = true
	}

	// Main function
	var err error
	load := loader.NewLoader(c)

	if flags.NArg() > 0 {
		err = load.LoadFile(flags.Arg(0))
	} else {
		err = load.LoadAllFile(".")
	}

	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	issues, err := detector.Detect(load.ListMap)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	printer.Print(issues, cli.outStream, cli.errStream)

	return ExitCodeOK
}
