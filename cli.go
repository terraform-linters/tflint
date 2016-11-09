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
		version      bool
		help         bool
		debug        bool
		format       string
		ignoreModule string
		ignoreRule   string
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	// Do not print default usage message
	flags.SetOutput(new(bytes.Buffer))

	flags.BoolVar(&version, "version", false, "Print version information and quit.")
	flags.BoolVar(&version, "v", false, "Alias for -version")
	flags.BoolVar(&help, "help", false, "Show usage (this page)")
	flags.BoolVar(&help, "h", false, "Alias for --help")
	flags.BoolVar(&debug, "debug", false, "Enable debug mode")
	flags.BoolVar(&debug, "d", false, "Alias for --debug")
	flags.StringVar(&format, "format", "default", "Specify output format")
	flags.StringVar(&format, "f", "default", "Alias for --format")
	flags.StringVar(&ignoreModule, "ignore-module", "", "Ignore specified module source")
	flags.StringVar(&ignoreRule, "ignore-rule", "", "Ignore specified rules")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(cli.errStream, "ERROR: `%s` is unknown options. Please run `tflint --help`\n", args[1])
		return ExitCodeError
	}
	if !printer.ValidateFormat(format) {
		fmt.Fprintf(cli.errStream, "ERROR: `%s` is unknown format. Please run `tflint --help`\n", format)
		return ExitCodeError
	}

	c := config.Init(ignoreModule, ignoreRule)

	// Show version
	if version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	if help {
		fmt.Fprintln(cli.outStream, `TFLint is a linter of Terraform.

Usage: tflint [<options>] <args>

Available options:
	-h, --help				show usage of TFLint. This page.
	-v, --version				print version information.
	-f, --format <format>			choose output format from "default" or "json"
	--ignore-module <source1,source2...>	ignore module by specified source.
	--ignore-rule <rule1,rule2...>		ignore rules.
	-d, --debug				enable debug mode.

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
	l := loader.NewLoader(c)
	if flags.NArg() > 0 {
		err = l.LoadFile(flags.Arg(0))
	} else {
		err = l.LoadAllFile(".")
	}
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}

	d, err := detector.NewDetector(l.ListMap, c)
	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	issues := d.Detect()

	p := printer.NewPrinter(cli.outStream, cli.errStream)
	p.Print(issues, format)

	return ExitCodeOK
}
