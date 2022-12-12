package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/terraform-linters/tflint/langserver"
)

func (cli *CLI) startLanguageServer(opts Options) int {
	if opts.Chdir != "" {
		fmt.Fprintf(cli.errStream, "Cannot use --chdir with --langserver\n")
		return ExitCodeError
	}
	configPath := opts.Config
	cliConfig := opts.toConfig()

	log.Println("Starting language server...")

	handler, plugin, err := langserver.NewHandler(configPath, cliConfig)
	if err != nil {
		log.Printf("Failed to start language server: %s", err)
		return ExitCodeError
	}
	if plugin != nil {
		defer plugin.Clean()
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
