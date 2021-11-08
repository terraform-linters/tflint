package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/logutils"
	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/terraform-linters/tflint/langserver"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

type jsonrpcMessage struct {
	ID      int         `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	JSONRPC string      `json:"jsonrpc"`
}

func TestMain(m *testing.M) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(""),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	os.Exit(m.Run())
}

func startServer(t *testing.T, configPath string) (io.Writer, io.Reader, *plugin.Plugin) {
	handler, plugin, err := langserver.NewHandler(configPath, tflint.EmptyConfig())
	if err != nil {
		t.Fatal(err)
	}

	stdin, stdinWriter := io.Pipe()
	stdoutReader, stdout := io.Pipe()

	var connOpt []jsonrpc2.ConnOpt
	jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(langserver.NewConn(stdin, stdout), jsonrpc2.VSCodeObjectCodec{}),
		handler,
		connOpt...,
	)

	return stdinWriter, stdoutReader, plugin
}

func pathToURI(path string) lsp.DocumentURI {
	path = filepath.ToSlash(path)
	parts := strings.SplitN(path, "/", 2)

	head := parts[0]
	if head != "" {
		head = "/" + head
	}

	rest := ""
	if len(parts) > 1 {
		rest = "/" + parts[1]
	}

	return lsp.DocumentURI("file://" + head + rest)
}

func shutdownRequest() string {
	return toJSONRPC2(`{"id":0,"method":"shutdown","params":{},"jsonrpc":"2.0"}`)
}

func exitRequest() string {
	return toJSONRPC2(`{"id":0,"method":"exit","params":{},"jsonrpc":"2.0"}`)
}

func emptyResponse() string {
	return toJSONRPC2(`{"id":0,"result":null,"jsonrpc":"2.0"}`)
}

func toJSONRPC2(json string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(json), json)
}

func withinFixtureDir(t *testing.T, dir string, test func(dir string)) {
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(current); err != nil {
			t.Fatal(err)
		}
	}()

	dir, err = filepath.Abs("test-fixtures/" + dir)
	if err != nil {
		t.Fatal(err)
	}
	test(dir)
}

func withinTempDir(t *testing.T, test func(dir string)) {
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(current); err != nil {
			t.Fatal(err)
		}
	}()

	dir, err := os.MkdirTemp("", "withinTempDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	test(dir)
}
