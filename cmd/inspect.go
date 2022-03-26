package cmd

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/rules"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) inspect(opts Options, dir string, filterFiles []string) int {
	// Setup config
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to load TFLint config; %w", err), map[string][]byte{})
		return ExitCodeError
	}
	if len(opts.Only) > 0 {
		for _, rule := range cfg.Rules {
			rule.Enabled = false
		}
	}
	cfg = cfg.Merge(opts.toConfig())

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

	rulesets := []tflint.RuleSet{&rules.RuleSet{}}
	config := cfg.ToPluginConfig()
	for name, ruleset := range rulesetPlugin.RuleSets {
		if err := ruleset.ApplyGlobalConfig(config); err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to apply global config to `%s` plugin; %w", name, err), cli.loader.Sources())
			return ExitCodeError
		}
		configSchema, err := ruleset.ConfigSchema()
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to fetch config schema from `%s` plugin; %w", name, err), cli.loader.Sources())
			return ExitCodeError
		}
		content, diags := cfg.PluginContent(name, configSchema)
		if diags.HasErrors() {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to parse `%s` plugin config; %w", name, diags), cli.loader.Sources())
			return ExitCodeError
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
	for _, rule := range rules.NewRules(cfg) {
		for _, runner := range runners {
			err := rule.Check(runner)
			if err != nil {
				cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to check `%s` rule; %w", rule.Name(), err), cli.loader.Sources())
				return ExitCodeError
			}
		}
	}

	for _, ruleset := range rulesetPlugin.RuleSets {
		for _, runner := range runners {
			err = ruleset.Check(plugin.NewGRPCServer(runner, rootRunner, cli.loader.Sources()))
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
	files, err := cli.loader.Files()
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to parse files; %w", err)
	}
	annotations, err := cli.loader.LoadAnnotations(dir)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load configuration tokens; %w", err)
	}
	variables, err := cli.loader.LoadValuesFiles(cfg.Varfiles...)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load values files; %w", err)
	}
	cliVars, err := tflint.ParseTFVariables(cfg.Variables, configs.Module.Variables)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to parse variables; %w", err)
	}
	variables = append(variables, cliVars)

	runner, err := tflint.NewRunner(cfg, files, annotations, configs, variables...)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to initialize a runner; %w", err)
	}

	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to prepare rule checking; %w", err)
	}

	return append(runners, runner), nil
}
