package cmd

import (
	"fmt"

	"github.com/spf13/afero"
	tfplugin "github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/rules"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) inspect(opts Options, dir string, filterFiles []string) int {
	// Setup config
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to load TFLint config", err), map[string][]byte{})
		return ExitCodeError
	}
	cfg = cfg.Merge(opts.toConfig())

	// Setup loader
	if !cli.testMode {
		cli.loader, err = tflint.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, cfg)
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to prepare loading", err), map[string][]byte{})
			return ExitCodeError
		}
	}

	// Setup runners
	runners, appErr := cli.setupRunners(opts, cfg, dir)
	if appErr != nil {
		cli.formatter.Print(tflint.Issues{}, appErr, cli.loader.Sources())
		return ExitCodeError
	}

	// Lookup plugins and validation
	plugin, err := tfplugin.Discovery(cfg)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to initialize plugins", err), cli.loader.Sources())
		return ExitCodeError
	}
	defer plugin.Clean()

	rulesets := []tflint.RuleSet{&rules.RuleSet{}}
	for _, ruleset := range plugin.RuleSets {
		rulesets = append(rulesets, ruleset)
	}
	if err := cfg.ValidateRules(rulesets...); err != nil {
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to check rule config", err), cli.loader.Sources())
		return ExitCodeError
	}

	// Run inspection
	for _, rule := range rules.NewRules(cfg) {
		for _, runner := range runners {
			err := rule.Check(runner)
			if err != nil {
				cli.formatter.Print(tflint.Issues{}, tflint.NewContextError(fmt.Sprintf("Failed to check `%s` rule", rule.Name()), err), cli.loader.Sources())
				return ExitCodeError
			}
		}
	}

	for _, ruleset := range plugin.RuleSets {
		err = ruleset.ApplyConfig(cfg.ToPluginConfig())
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to apply config to plugins", err), cli.loader.Sources())
		}
		for _, runner := range runners {
			err = ruleset.Check(tfplugin.NewServer(runner))
			if err != nil {
				cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to check ruleset", err), cli.loader.Sources())
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

func (cli *CLI) setupRunners(opts Options, cfg *tflint.Config, dir string) ([]*tflint.Runner, *tflint.Error) {
	configs, err := cli.loader.LoadConfig(dir)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to load configurations", err)
	}
	annotations, err := cli.loader.LoadAnnotations(dir)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to load configuration tokens", err)
	}
	variables, err := cli.loader.LoadValuesFiles(cfg.Varfiles...)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to load values files", err)
	}
	cliVars, err := tflint.ParseTFVariables(cfg.Variables, configs.Module.Variables)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to parse variables", err)
	}
	variables = append(variables, cliVars)

	runner, err := tflint.NewRunner(cfg, annotations, configs, variables...)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to initialize a runner", err)
	}

	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to prepare rule checking", err)
	}

	return append(runners, runner), nil
}
