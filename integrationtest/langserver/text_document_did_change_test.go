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

func Test_textDocumentDidChange(t *testing.T) {
	withinFixtureDir(t, "workdir", func(dir string) {
		src, err := os.ReadFile(dir + "/main.tf")
		if err != nil {
			t.Fatal(err)
		}
		uri := pathToURI(dir + "/main.tf")

		stdin, stdout, plugin := startServer(t, dir+"/.tflint.hcl")
		defer plugin.Clean()

		req, err := json.Marshal(jsonrpcMessage{
			ID:     0,
			Method: "textDocument/didChange",
			Params: lsp.DidChangeTextDocumentParams{
				TextDocument: lsp.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: lsp.TextDocumentIdentifier{
						URI: uri,
					},
					Version: 2,
				},
				ContentChanges: []lsp.TextDocumentContentChangeEvent{
					{
						Text: `
resource "aws_instance" "foo" {
	ami = "ami-12345678"
}`,
					},
				},
			},
			JSONRPC: "2.0",
		})
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			fmt.Fprint(stdin, initializeRequest())
			fmt.Fprint(stdin, didOpenRequest(uri, string(src), t))
			fmt.Fprint(stdin, toJSONRPC2(string(req)))
			fmt.Fprint(stdin, shutdownRequest())
			fmt.Fprint(stdin, exitRequest())
		}()

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(stdout); err != nil {
			t.Fatal(err)
		}

		expected := initializeResponse() + didOpenResponse(uri, t) + noDiagnosticsResponse(uri, t) + emptyResponse()
		if !cmp.Equal(expected, buf.String()) {
			t.Fatalf("Diff: %s", cmp.Diff(expected, buf.String()))
		}

		// Assert no changes for actual files
		changedSrc, err := os.ReadFile(dir + "/main.tf")
		if err != nil {
			t.Fatal(err)
		}
		if string(src) != string(changedSrc) {
			t.Fatal("textDocument/didChange event is rewriting the actual files")
		}
	})
}

func noDiagnosticsResponse(uri lsp.DocumentURI, t *testing.T) string {
	didChangeResponse, err := json.Marshal(jsonrpcMessage{
		Method: "textDocument/publishDiagnostics",
		Params: lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []lsp.Diagnostic{},
		},
		JSONRPC: "2.0",
	})
	if err != nil {
		t.Fatal(err)
	}

	return toJSONRPC2(string(didChangeResponse))
}
