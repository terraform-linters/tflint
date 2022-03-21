package langserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	lsp "github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/rules"
	"github.com/terraform-linters/tflint/tflint"
)

// NewHandler returns a new JSON-RPC handler
func NewHandler(configPath string, cliConfig *tflint.Config) (jsonrpc2.Handler, *plugin.Plugin, error) {
	cfg, err := tflint.LoadConfig(configPath)
	if err != nil {
		return nil, nil, err
	}
	if cliConfig.DisabledByDefault {
		for _, rule := range cfg.Rules {
			rule.Enabled = false
		}
	}
	cfg = cfg.Merge(cliConfig)

	rulsetPlugin, err := plugin.Discovery(cfg)
	if err != nil {
		return nil, nil, err
	}

	rulesets := []tflint.RuleSet{&rules.RuleSet{}}
	for _, ruleset := range rulsetPlugin.RuleSets {
		rulesets = append(rulesets, ruleset)
	}
	if err := cliConfig.ValidateRules(rulesets...); err != nil {
		return nil, nil, err
	}

	return jsonrpc2.HandlerWithError((&handler{
		configPath: configPath,
		cliConfig:  cliConfig,
		config:     cfg,
		fs:         afero.NewCopyOnWriteFs(afero.NewOsFs(), afero.NewMemMapFs()),
		rules:      rules.NewRules(cfg),
		plugin:     rulsetPlugin,
		diagsPaths: []string{},
	}).handle), rulsetPlugin, nil
}

type handler struct {
	configPath string
	cliConfig  *tflint.Config
	config     *tflint.Config
	fs         afero.Fs
	rootDir    string
	rules      []rules.Rule
	plugin     *plugin.Plugin
	shutdown   bool
	diagsPaths []string
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params != nil {
		params, err := json.Marshal(&req.Params)
		if err != nil {
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeParseError,
				Message: err.Error(),
				Data:    req.Params,
			}
		}
		log.Println(fmt.Sprintf("Received `%s` with `%s`", req.Method, string(params)))
	} else {
		log.Println(fmt.Sprintf("Received `%s`", req.Method))
	}

	if h.shutdown && req.Method != "exit" {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidRequest,
			Message: "server is shutting down",
		}
	}

	switch req.Method {
	case "initialize":
		return initialize(ctx, conn, req)
	case "initialized":
		return nil, nil
	case "shutdown":
		h.shutdown = true
		return nil, nil
	case "exit":
		return nil, conn.Close()
	case "textDocument/didOpen":
		return h.textDocumentDidOpen(ctx, conn, req)
	case "textDocument/didClose":
		return nil, nil
	case "textDocument/didChange":
		return h.textDocumentDidChange(ctx, conn, req)
	case "workspace/didChangeWatchedFiles":
		return h.workspaceDidChangeWatchedFiles(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.CodeMethodNotFound,
		Message: fmt.Sprintf("unsupported request: %s", req.Method),
	}
}

func (h *handler) chdir(dir string) error {
	if h.rootDir != dir {
		log.Println(fmt.Sprintf("Changing directory: %s", dir))
		if err := os.Chdir(dir); err != nil {
			return fmt.Errorf("Failed to chdir to %s: %s", dir, err)
		}
		h.rootDir = dir
	}
	return nil
}

func (h *handler) inspect() (map[string][]lsp.Diagnostic, error) {
	ret := map[string][]lsp.Diagnostic{}

	loader, err := tflint.NewLoader(afero.Afero{Fs: h.fs}, h.config)
	if err != nil {
		return ret, fmt.Errorf("Failed to prepare loading: %w", err)
	}

	configs, err := loader.LoadConfig(".")
	if err != nil {
		return ret, fmt.Errorf("Failed to load configurations: %w", err)
	}
	files, err := loader.Files()
	if err != nil {
		return ret, fmt.Errorf("Failed to parse files; %w", err)
	}
	annotations, err := loader.LoadAnnotations(".")
	if err != nil {
		return ret, fmt.Errorf("Failed to load configuration tokens: %w", err)
	}
	variables, err := loader.LoadValuesFiles(h.config.Varfiles...)
	if err != nil {
		return ret, fmt.Errorf("Failed to load values files: %w", err)
	}
	cliVars, err := tflint.ParseTFVariables(h.config.Variables, configs.Module.Variables)
	if err != nil {
		return ret, fmt.Errorf("Failed to parse variables: %w", err)
	}
	variables = append(variables, cliVars)

	runner, err := tflint.NewRunner(h.config, files, annotations, configs, variables...)
	if err != nil {
		return ret, fmt.Errorf("Failed to initialize a runner: %w", err)
	}
	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return ret, fmt.Errorf("Failed to prepare rule checking: %w", err)
	}
	runners = append(runners, runner)

	for _, rule := range h.rules {
		for _, runner := range runners {
			err := rule.Check(runner)
			if err != nil {
				return ret, fmt.Errorf("Failed to check `%s` rule: %w", rule.Name(), err)
			}
		}
	}

	config := h.config.ToPluginConfig()
	for name, ruleset := range h.plugin.RuleSets {
		if err := ruleset.ApplyGlobalConfig(config); err != nil {
			return ret, fmt.Errorf("Failed to apply global config to `%s` plugin", name)
		}
		configSchema, err := ruleset.ConfigSchema()
		if err != nil {
			return ret, fmt.Errorf("Failed to fetch config schema from `%s` plugin", name)
		}
		content, diags := h.config.PluginContent(name, configSchema)
		if diags.HasErrors() {
			return ret, fmt.Errorf("Failed to parse `%s` plugin config", name)
		}
		err = ruleset.ApplyConfig(content, h.config.Sources())
		if err != nil {
			return ret, fmt.Errorf("Failed to apply config to `%s` plugin", name)
		}
		for _, runner := range runners {
			err = ruleset.Check(plugin.NewGRPCServer(runner, runners[len(runners)-1], loader.Sources()))
			if err != nil {
				return ret, fmt.Errorf("Failed to check ruleset: %w", err)
			}
		}
	}

	// In order to publish that the issue has been fixed,
	// notify also the path where the past diagnostics were published.
	for _, path := range h.diagsPaths {
		ret[path] = []lsp.Diagnostic{}
	}
	h.diagsPaths = []string{}

	for _, runner := range runners {
		for _, issue := range runner.LookupIssues() {
			path := filepath.Join(h.rootDir, issue.Range.Filename)
			h.diagsPaths = append(h.diagsPaths, path)

			diag := lsp.Diagnostic{
				Message:  issue.Message,
				Severity: toLSPSeverity(issue.Rule.Severity()),
				Range: lsp.Range{
					Start: lsp.Position{Line: issue.Range.Start.Line - 1, Character: issue.Range.Start.Column - 1},
					End:   lsp.Position{Line: issue.Range.End.Line - 1, Character: issue.Range.End.Column - 1},
				},
			}

			if ret[path] == nil {
				ret[path] = []lsp.Diagnostic{diag}
			} else {
				ret[path] = append(ret[path], diag)
			}
		}
	}

	return ret, nil
}

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
	case tflint.ERROR:
		return lsp.Error
	case tflint.WARNING:
		return lsp.Warning
	case tflint.NOTICE:
		return lsp.Information
	default:
		panic(fmt.Sprintf("Unexpected severity: %s", severity))
	}
}
