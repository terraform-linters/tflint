package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/k0kubun/pp"
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

	if flags.NArg() > 0 {
		b, err := ioutil.ReadFile(flags.Arg(0))
		if err != nil {
			fmt.Fprintf(cli.errStream, "ERROR: Cannot open file %s\n", flags.Arg(0))
			return ExitCodeError
		}
		root, err := parser.Parse(b)
		if err != nil {
			fmt.Fprintf(cli.errStream, "ERROR: Parse error occurred by %s\n", flags.Arg(0))
			return ExitCodeError
		}

		validInstanceType := map[string]bool{
			"t2.nano":     true,
			"t2.micro":    true,
			"t2.small":    true,
			"t2.medium":   true,
			"t2.large":    true,
			"m4.large":    true,
			"m4.xlarge":   true,
			"m4.2xlarge":  true,
			"m4.4xlarge":  true,
			"m4.10xlarge": true,
			"m4.16xlarge": true,
			"m3.medium":   true,
			"m3.large":    true,
			"m3.xlarge":   true,
			"m3.2xlarge":  true,
			"c4.large":    true,
			"c4.2xlarge":  true,
			"c4.4xlarge":  true,
			"c4.8xlarge":  true,
			"c3.large":    true,
			"c3.xlarge":   true,
			"c3.2xlarge":  true,
			"c3.4xlarge":  true,
			"c3.8xlarge":  true,
			"x1.32xlarge": true,
			"r3.large":    true,
			"r3.xlarge":   true,
			"r3.2xlarge":  true,
			"r3.4xlarge":  true,
			"r3.8xlarge":  true,
			"p2.xlarge":   true,
			"p2.8xlarge":  true,
			"p2.16xlarge": true,
			"g2.2xlarge":  true,
			"g2.8xlarge":  true,
			"i2.xlarge":   true,
			"i2.2xlarge":  true,
			"i2.4xlarge":  true,
			"i2.8xlarge":  true,
			"d2.xlarge":   true,
			"d2.2xlarge":  true,
			"d2.4xlarge":  true,
			"d2.8xlarge":  true,
		}

		validPreviousGenerationInstanceType := map[string]bool{
			"t1.micro":    true,
			"m1.small":    true,
			"m1.medium":   true,
			"m1.large":    true,
			"m1.xlarge":   true,
			"c1.medium":   true,
			"c1.xlarge":   true,
			"cc2.8xlarge": true,
			"cg1.4xlarge": true,
			"m2.xlarge":   true,
			"m2.2xlarge":  true,
			"m2.4xlarge":  true,
			"cr1.8xlarge": true,
			"hi1.4xlarge": true,
			"hs1.8xlarge": true,
		}

		list, _ := root.Node.(*ast.ObjectList)
		for _, item := range list.Filter("resource", "aws_instance").Items {
			instanceTypeToken := item.Val.(*ast.ObjectType).List.Filter("instance_type").Items[0].Val.(*ast.LiteralType).Token
			instanceTypeKey := strings.Trim(instanceTypeToken.Text, "\"")

			if !validInstanceType[instanceTypeKey] && !validPreviousGenerationInstanceType[instanceTypeKey] {
				fmt.Fprintf(cli.outStream, "WARNING: %s is invalid instance type. Line: %d in %s\n", instanceTypeToken.Text, instanceTypeToken.Pos.Line, flags.Arg(0))
			}

			if validPreviousGenerationInstanceType[instanceTypeKey] {
				fmt.Fprintf(cli.outStream, "NOTICE: %s is previous generation instance type. Line: %d in %s\n", instanceTypeToken.Text, instanceTypeToken.Pos.Line, flags.Arg(0))
			}

			pp.Print(instanceTypeToken)
		}
	}

	return ExitCodeOK
}
