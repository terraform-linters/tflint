package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/hcl/ast"
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
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.BoolVar(&version, "version", false, "Print version information and quit.")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintln(cli.errStream, "ERROR: Parse error occurred.\n")
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Main function
	var listmap map[string]*ast.ObjectList
	var err error
	if flags.NArg() > 0 {
		listmap, err = loader.LoadFile(nil, flags.Arg(0))
	} else {
		listmap, err = loader.LoadAllFile()
	}

	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return ExitCodeError
	}
	issues := detector.Detect(listmap)
	printer.Print(issues, cli.outStream, cli.errStream)

	return ExitCodeOK
}
