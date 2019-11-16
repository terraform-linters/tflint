package langserver

import (
	"context"
	"fmt"
	"log"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/rules"
	"github.com/terraform-linters/tflint/tflint"
)

func (h *handler) workspaceDidChangeWatchedFiles(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if h.rootDir == "" {
		return nil, fmt.Errorf("root directory is undefined")
	}

	newConfig, err := tflint.LoadConfig(h.configPath)
	if err != nil {
		return nil, err
	}
	h.config = newConfig.Merge(h.cliConfig)
	h.rules = rules.NewRules(h.config)

	h.fs = afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs())

	diagnostics, err := h.inspect()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("Notify `textDocument/publishDiagnostics` with `%#v`", diagnostics))
	for path, diags := range diagnostics {
		err = conn.Notify(
			ctx,
			"textDocument/publishDiagnostics",
			lsp.PublishDiagnosticsParams{
				URI:         pathToURI(path),
				Diagnostics: diags,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("Failed to notify `textDocument/publishDiagnostics`: %s", err)
		}
	}

	return nil, nil
}
