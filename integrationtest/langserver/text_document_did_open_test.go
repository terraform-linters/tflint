package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	lsp "github.com/sourcegraph/go-lsp"
)

func Test_textDocumentDidOpen(t *testing.T) {
	withinFixtureDir(t, "workdir", func(dir string) {
		src, err := os.ReadFile(dir + "/main.tf")
		if err != nil {
			t.Fatal(err)
		}
		uri := pathToURI(dir + "/main.tf")

		stdin, stdout, plugin := startServer(t, dir+"/.tflint.hcl")
		defer plugin.Clean()

		go func() {
			fmt.Fprint(stdin, initializeRequest())
			fmt.Fprint(stdin, didOpenRequest(uri, string(src), t))
			fmt.Fprint(stdin, shutdownRequest())
			fmt.Fprint(stdin, exitRequest())
		}()

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(stdout); err != nil {
			t.Fatal(err)
		}

		expected := initializeResponse() + didOpenResponse(uri, t) + emptyResponse()
		if !cmp.Equal(expected, buf.String()) {
			t.Fatalf("Diff: %s", cmp.Diff(expected, buf.String()))
		}
	})
}

func didOpenRequest(uri lsp.DocumentURI, src string, t *testing.T) string {
	req, err := json.Marshal(jsonrpcMessage{
		ID:     0,
		Method: "textDocument/didOpen",
		Params: lsp.DidOpenTextDocumentParams{
			TextDocument: lsp.TextDocumentItem{
				URI:        uri,
				LanguageID: "terraform",
				Version:    1,
				Text:       src,
			},
		},
		JSONRPC: "2.0",
	})
	if err != nil {
		t.Fatal(err)
	}

	return toJSONRPC2(string(req))
}

func didOpenResponse(uri lsp.DocumentURI, t *testing.T) string {
	res, err := json.Marshal(jsonrpcMessage{
		Method: "textDocument/publishDiagnostics",
		Params: lsp.PublishDiagnosticsParams{
			URI: uri,
			Diagnostics: []lsp.Diagnostic{
				{
					Message:  `instance type is t1.2xlarge`,
					Severity: lsp.Error,
					Range: lsp.Range{
						Start: lsp.Position{Line: 1, Character: 20},
						End:   lsp.Position{Line: 1, Character: 32},
					},
				},
			},
		},
		JSONRPC: "2.0",
	})
	if err != nil {
		t.Fatal(err)
	}

	return toJSONRPC2(string(res))
}

func Test_textDocumentDidOpen_pathFunctions(t *testing.T) {
	withinFixtureDir(t, "path_functions", func(dir string) {
		src, err := os.ReadFile(dir + "/main.tf")
		if err != nil {
			t.Fatal(err)
		}
		uri := pathToURI(dir + "/main.tf")

		stdin, stdout, plugin := startServer(t, dir+"/.tflint.hcl")
		defer plugin.Clean()

		go func() {
			fmt.Fprint(stdin, initializeRequest())
			fmt.Fprint(stdin, didOpenRequest(uri, string(src), t))
			fmt.Fprint(stdin, shutdownRequest())
			fmt.Fprint(stdin, exitRequest())
		}()

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(stdout); err != nil {
			t.Fatal(err)
		}

		didOpenResponse, err := json.Marshal(jsonrpcMessage{
			Method: "textDocument/publishDiagnostics",
			Params: lsp.PublishDiagnosticsParams{
				URI: uri,
				Diagnostics: []lsp.Diagnostic{
					{
						Message:  `instance type is t1.2xlarge`,
						Severity: lsp.Error,
						Range: lsp.Range{
							Start: lsp.Position{Line: 1, Character: 20},
							End:   lsp.Position{Line: 1, Character: 53},
						},
					},
				},
			},
			JSONRPC: "2.0",
		})
		if err != nil {
			t.Fatal(err)
		}

		expected := initializeResponse() + toJSONRPC2(string(didOpenResponse)) + emptyResponse()
		if !cmp.Equal(expected, buf.String()) {
			t.Fatalf("Diff: %s", cmp.Diff(expected, buf.String()))
		}
	})
}
