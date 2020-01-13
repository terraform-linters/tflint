package langserver

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/logutils"
	"github.com/sourcegraph/jsonrpc2"
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
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(""),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	os.Exit(m.Run())
}

func startServer(t *testing.T) (io.Writer, io.Reader) {
	handler, _, err := NewHandler(".tflint.hcl", tflint.EmptyConfig())
	if err != nil {
		t.Fatal(err)
	}

	stdin, stdinWriter := io.Pipe()
	stdoutReader, stdout := io.Pipe()

	var connOpt []jsonrpc2.ConnOpt
	jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(NewConn(stdin, stdout), jsonrpc2.VSCodeObjectCodec{}),
		handler,
		connOpt...,
	)

	return stdinWriter, stdoutReader
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
	defer os.Chdir(current)

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
	defer os.Chdir(current)

	dir, err := ioutil.TempDir("", "withinTempDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	test(dir)
}
