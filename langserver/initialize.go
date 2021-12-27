package langserver

import (
	lsp "github.com/sourcegraph/go-lsp"
)

func initialize() (result interface{}, err error) {
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
