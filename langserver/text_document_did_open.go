package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/afero"
)

func (h *handler) textDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidRequest,
			Message: "request params are nil",
		}
	}

	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: err.Error(),
			Data:    req.Params,
		}
	}

	openedPath, err := uriToPath(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	if err := h.chdir(filepath.Dir(openedPath)); err != nil {
		return nil, err
	}

	if err := afero.WriteFile(h.fs, filepath.Base(openedPath), []byte(params.TextDocument.Text), os.ModePerm); err != nil {
		return nil, fmt.Errorf("Failed to synchronize TextDocument.Text: %s", err)
	}

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
