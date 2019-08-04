package langserver

import (
	"context"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func initialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	return lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
				},
			},
		},
	}, nil
}
