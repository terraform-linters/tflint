package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	lsp "github.com/sourcegraph/go-lsp"
)

func Test_workspaceDidChangeWatchedFiles(t *testing.T) {
	withinTempDir(t, func(dir string) {
		content := `resource "aws_instance" "foo" {
    instance_type = "t1.2xlarge"
}`

		config := `
plugin "testing" {
    enabled = true
}`

		changedConfig := `
plugin "testing" {
    enabled = true
}

rule "aws_instance_example_type" {
    enabled = false
}`

		if err := os.WriteFile(dir+"/main.tf", []byte(content), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dir+"/.tflint.hcl", []byte(config), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		uri := pathToURI(dir + "/main.tf")

		stdin, stdout, plugin := startServer(t, dir+"/.tflint.hcl")
		defer plugin.Clean()

		req, err := json.Marshal(jsonrpcMessage{
			ID:     0,
			Method: "workspace/didChangeWatchedFiles",
			Params: lsp.DidChangeWatchedFilesParams{
				Changes: []lsp.FileEvent{
					{
						URI:  lsp.DocumentURI(dir + "/.tflint.hcl"),
						Type: int(lsp.Created),
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
			fmt.Fprint(stdin, didOpenRequest(uri, content, t))
			// Change config file from outside of LSP
			_ = os.WriteFile(dir+"/.tflint.hcl", []byte(changedConfig), os.ModePerm)
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
	})
}

func Test_workspaceDidChangeWatchedFiles_withDeletedFile(t *testing.T) {
	withinTempDir(t, func(dir string) {
		content := `
resource "aws_instance" "foo" {
    instance_type = var.instance_type
}
variable "instance_type" {}
`

		if err := os.WriteFile(dir+"/main.tf", []byte(content), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		uri := pathToURI(dir + "/main.tf")

		valueFile := `
instance_type = "t1.2xlarge"
`
		if err := os.WriteFile(dir+"/terraform.tfvars", []byte(valueFile), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		config := `
plugin "testing" {
    enabled = true
}`
		if err := os.WriteFile(dir+"/.tflint.hcl", []byte(config), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		stdin, stdout, plugin := startServer(t, dir+"/.tflint.hcl")
		defer plugin.Clean()

		req, err := json.Marshal(jsonrpcMessage{
			ID:     0,
			Method: "workspace/didChangeWatchedFiles",
			Params: lsp.DidChangeWatchedFilesParams{
				Changes: []lsp.FileEvent{
					{
						URI:  lsp.DocumentURI(dir + "/terraform.tfvars"),
						Type: int(lsp.Deleted),
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
			fmt.Fprint(stdin, didOpenRequest(uri, content, t))
			// Wait didOpen inspection
			time.Sleep(100 * time.Millisecond)
			// Remove values file from outside of LSP
			os.Remove(dir + "/terraform.tfvars")
			fmt.Fprint(stdin, toJSONRPC2(string(req)))
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
							Start: lsp.Position{Line: 2, Character: 20},
							End:   lsp.Position{Line: 2, Character: 37},
						},
					},
				},
			},
			JSONRPC: "2.0",
		})
		if err != nil {
			t.Fatal(err)
		}

		expected := initializeResponse() + toJSONRPC2(string(didOpenResponse)) + noDiagnosticsResponse(uri, t) + emptyResponse()
		if !cmp.Equal(expected, buf.String()) {
			t.Fatalf("Diff: %s", cmp.Diff(expected, buf.String()))
		}
	})
}
