package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/wata727/tflint/rules"
	"github.com/wata727/tflint/tflint"
)

func NewHandler(config *tflint.Config) (jsonrpc2.Handler, error) {
	return jsonrpc2.HandlerWithError((&handler{
		config: config,
		rules:  rules.NewRules(config),
	}).handle), nil
}

type handler struct {
	config *tflint.Config
	rules  []rules.Rule
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params != nil {
		params, err := json.Marshal(&req.Params)
		if err != nil {
			return nil, err
		}
		log.Println(fmt.Sprintf("Received `%s` with `%s`", req.Method, string(params)))
	} else {
		log.Println(fmt.Sprintf("Received `%s`"))
	}

	switch req.Method {
	case "initialize":
		return initialize(ctx, conn, req)
	case "shutdown":
		// TODO
	case "exit":
		return nil, conn.Close()
	case "textDocument/didOpen":
		// TODO
	case "textDocument/didChange":
		return h.textDocumentDidChange(ctx, conn, req)
	}

	log.Println(fmt.Sprintf("unsupported request: %s", req.Method))
	return nil, nil
}

func uriToPath(uri lsp.DocumentURI) string {
	url, err := url.Parse(string(uri))
	if err != nil {
		panic(err)
	}
	return url.Path
}
