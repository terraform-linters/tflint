package langserver

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	lsp "github.com/sourcegraph/go-lsp"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

func uriToPath(uri lsp.DocumentURI) (string, error) {
	uriToReplace, err := url.QueryUnescape(string(uri))
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		return strings.Replace(uriToReplace, "file:///", "", 1), nil
	}
	return strings.Replace(uriToReplace, "file://", "", 1), nil
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

func toLSPSeverity(severity tflint.Severity) lsp.DiagnosticSeverity {
	switch severity {
	case sdk.ERROR:
		return lsp.Error
	case sdk.WARNING:
		return lsp.Warning
	case sdk.NOTICE:
		return lsp.Information
	default:
		panic(fmt.Sprintf("Unexpected severity: %s", severity))
	}
}
