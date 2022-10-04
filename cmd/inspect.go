package cmd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cli *CLI) inspect(opts Options, dir string, filterFiles []string) int {
	// Setup config
	cfg, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to load TFLint config; %w", err), map[string][]byte{})
		return ExitCodeError
	}
	// tflint-plugin-sdk v0.13+ doesn't need to disable rules config when enabling the only option.
	// This is for the backward compatibility.
	if len(opts.Only) > 0 {
		for _, rule := range cfg.Rules {
			rule.Enabled = false
		}
	}
	cfg.Merge(opts.toConfig())
	cli.formatter.Format = cfg.Format

	// Setup loader
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, cfg)
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to prepare loading; %w", err), map[string][]byte{})
			return ExitCodeError
		}
	}

	// Setup runners
	runners, appErr := cli.setupRunners(opts, cfg, dir)
	if appErr != nil {
		cli.formatter.Print(tflint.Issues{}, appErr, cli.loader.Sources())
		return ExitCodeError
	}
	rootRunner := runners[len(runners)-1]

	// Lookup plugins and validation
	rulesetPlugin, err := plugin.Discovery(cfg)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to initialize plugins; %w", err), cli.loader.Sources())
		return ExitCodeError
	}
	defer rulesetPlugin.Clean()

	rulesets := []tflint.RuleSet{}
	config := cfg.ToPluginConfig()
	for name, ruleset := range rulesetPlugin.RuleSets {
		constraints, err := ruleset.VersionConstraints()
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
				// VersionConstraints endpoint is available in tflint-plugin-sdk v0.14+.
				// Skip verification if not available.
			} else {
				cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to get TFLint version constraints to `%s` plugin; %w", name, err), cli.loader.Sources())
				return ExitCodeError
			}
		}
		if !constraints.Check(tflint.Version) {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to satisfy version constraints; tflint-ruleset-%s requires %s, but TFLint version is %s", name, constraints, tflint.Version), cli.loader.Sources())
			return ExitCodeError
		}

		if err := ruleset.ApplyGlobalConfig(config); err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to apply global config to `%s` plugin; %w", name, err), cli.loader.Sources())
			return ExitCodeError
		}
		configSchema, err := ruleset.ConfigSchema()
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to fetch config schema from `%s` plugin; %w", name, err), cli.loader.Sources())
			return ExitCodeError
		}
		content := &hclext.BodyContent{}
		if plugin, exists := cfg.Plugins[name]; exists {
			var diags hcl.Diagnostics
			content, diags = plugin.Content(configSchema)
			if diags.HasErrors() {
				cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to parse `%s` plugin config; %w", name, diags), cli.loader.Sources())
				return ExitCodeError
			}
		}
		err = ruleset.ApplyConfig(content, cfg.Sources())
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to apply config to `%s` plugin; %w", name, err), cli.loader.Sources())
			return ExitCodeError
		}

		rulesets = append(rulesets, ruleset)
	}
	if err := cfg.ValidateRules(rulesets...); err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to check rule config; %w", err), cli.loader.Sources())
		return ExitCodeError
	}

	// Run inspection
	for _, ruleset := range rulesetPlugin.RuleSets {
		for _, runner := range runners {
			err = ruleset.Check(plugin.NewGRPCServer(runner, rootRunner, cli.loader.Files()))
			if err != nil {
				cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to check ruleset; %w", err), cli.loader.Sources())
				return ExitCodeError
			}
		}
	}

	issues := tflint.Issues{}
	for _, runner := range runners {
		issues = append(issues, runner.LookupIssues(filterFiles...)...)
	}

	// Print issues
	cli.formatter.Print(issues, nil, cli.loader.Sources())

	if len(issues) > 0 && !cfg.Force {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) setupRunners(opts Options, cfg *tflint.Config, dir string) ([]*tflint.Runner, error) {
	configs, err := cli.loader.LoadConfig(dir)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load configurations; %w", err)
	}
	annotations, err := cli.loader.LoadAnnotations(dir)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load configuration tokens; %w", err)
	}
	variables, err := cli.loader.LoadValuesFiles(cfg.Varfiles...)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load values files; %w", err)
	}
	cliVars, diags := terraform.ParseVariableValues(cfg.Variables, configs.Module.Variables)
	if diags.HasErrors() {
		return []*tflint.Runner{}, fmt.Errorf("Failed to parse variables; %w", diags)
	}
	variables = append(variables, cliVars)

	runner, err := tflint.NewRunner(cfg, annotations, configs, variables...)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to initialize a runner; %w", err)
	}

	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to prepare rule checking; %w", err)
	}

	return append(runners, runner), nil
}
