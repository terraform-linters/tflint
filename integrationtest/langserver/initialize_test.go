package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_initialize(t *testing.T) {
	withinFixtureDir(t, "workdir", func(dir string) {
		stdin, stdout, plugin := startServer(t, dir+"/.tflint.hcl")
		defer plugin.Clean()

		go func() {
			fmt.Fprint(stdin, initializeRequest())
			fmt.Fprint(stdin, shutdownRequest())
			fmt.Fprint(stdin, exitRequest())
		}()

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(stdout); err != nil {
			t.Fatal(err)
		}

		expected := initializeResponse() + emptyResponse()
		if !cmp.Equal(expected, buf.String()) {
			t.Fatalf("Diff: %s", cmp.Diff(expected, buf.String()))
		}
	})
}

func initializeRequest() string {
	return toJSONRPC2(`{"id":0,"method":"initialize","params":{},"jsonrpc":"2.0"}`)
}

func initializeResponse() string {
	return toJSONRPC2(`{"id":0,"result":{"capabilities":{"textDocumentSync":{"openClose":true,"change":1}}},"jsonrpc":"2.0"}`)
}
