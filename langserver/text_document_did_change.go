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

func (h *handler) textDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidRequest,
			Message: "request params are nil",
		}
	}

	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeParseError,
			Message: err.Error(),
			Data:    req.Params,
		}
	}

	changedPath, err := uriToPath(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	if err := h.chdir(filepath.Dir(changedPath)); err != nil {
		return nil, err
	}

	for idx, contentChange := range params.ContentChanges {
		if err := afero.WriteFile(h.fs, filepath.Base(changedPath), []byte(contentChange.Text), os.ModePerm); err != nil {
			return nil, fmt.Errorf("Failed to synchronize contentChanges[%d].Text: %s", idx, err)
		}
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
