package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl2/hcl"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/wata727/tflint/tflint"
)

func (h *handler) textDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	changedPath := uriToPath(params.TextDocument.URI)
	if err := os.Chdir(filepath.Dir(changedPath)); err != nil {
		return nil, err
	}

	loader, err := tflint.NewLoader(h.config)
	if err != nil {
		return nil, err
	}

	// FIXME: Inspect the passed configuration file
	configs, err := loader.LoadConfig(".")
	if err != nil {
		return nil, err
	}
	annotations, err := loader.LoadAnnotations(".")
	if err != nil {
		return nil, err
	}
	variables, err := loader.LoadValuesFiles(h.config.Varfile...)
	if err != nil {
		return nil, err
	}
	cliVars, err := tflint.ParseTFVariables(h.config.Variables, configs.Module.Variables)
	if err != nil {
		return nil, err
	}
	variables = append(variables, cliVars)

	runner, err := tflint.NewRunner(h.config, annotations, configs, variables...)
	if err != nil {
		return nil, err
	}

	for _, rule := range h.rules {
		err := rule.Check(runner)
		if err != nil {
			return nil, err
		}
	}

	runner.WalkResourceAttributes("aws_instance", "instance_type", func(attr *hcl.Attribute) error {
		var ret string
		runner.EvaluateExpr(attr.Expr, &ret)
		log.Printf("instance_type: %#v", ret)
		return nil
	})

	diags := []lsp.Diagnostic{}
	for _, issue := range runner.LookupIssues() {
		if issue.File == filepath.Base(changedPath) {
			diags = append(diags, lsp.Diagnostic{
				Message:  issue.Message,
				Severity: lsp.Error,
				Range: lsp.Range{
					Start: lsp.Position{Line: issue.Line - 1, Character: 0},
					End:   lsp.Position{Line: issue.Line - 1, Character: 100},
				},
			})
		}
	}

	log.Println(fmt.Sprintf("Notify `textDocument/publishDiagnostics` with `%#v`", diags))
	conn.Notify(
		ctx,
		"textDocument/publishDiagnostics",
		lsp.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: diags,
		},
	)

	return nil, nil
}
