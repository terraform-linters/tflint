package cmd

import (
	"fmt"
	"log"

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
			cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to prepare loading", err), map[string][]byte{})
			return ExitCodeError
		}
	}

	// Setup runners
	runners, appErr := cli.setupRunners(cfg, dir)
	if appErr != nil {
		cli.formatter.Print(tflint.Issues{}, appErr, cli.loader.Sources())
		return ExitCodeError
	}
	rootRunner := runners[len(runners)-1]

	// AWS plugin is automatically enabled from your provider requirements, even if the plugin isn't explicitly enabled.
	if _, exists := cfg.Plugins["aws"]; !exists {
		reqs, diags := rootRunner.TFConfig.ProviderRequirements()
		if diags.HasErrors() {
			cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to get Terraform provider requirements", diags), cli.loader.Sources())
			return ExitCodeError
		}
		for addr := range reqs {
			if addr.Type == "aws" {
				log.Print("[INFO] AWS provider requirements found. Enable the plugin `aws` automatically")
				fmt.Fprintln(cli.errStream, "WARNING: The plugin `aws` is not explicitly enabled. The bundled plugin will be enabled instead, but it is deprecated and will be removed in a future version. Please see https://github.com/terraform-linters/tflint/pull/1160 for details.")
				cfg.Plugins["aws"] = &tflint.PluginConfig{
					Name:    "aws",
					Enabled: true,
				}
			}
		}
	}

	// Lookup plugins and validation
	plugin, err := tfplugin.Discovery(cfg)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to initialize plugins", err), cli.loader.Sources())
		return ExitCodeError
	}
	defer plugin.Clean()

	rulesets := []tflint.RuleSet{&rules.RuleSet{}}
	for name, ruleset := range plugin.RuleSets {
		err = ruleset.ApplyConfig(cfg.ToPluginConfig(name))
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to apply config to plugins", err), cli.loader.Sources())
			return ExitCodeError
		}
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
		for _, runner := range runners {
			err = ruleset.Check(tfplugin.NewServer(runner, rootRunner, cli.loader.Sources()))
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

func (cli *CLI) setupRunners(cfg *tflint.Config, dir string) ([]*tflint.Runner, *tflint.Error) {
	configs, err := cli.loader.LoadConfig(dir)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to load configurations", err)
	}
	files, err := cli.loader.Files()
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to parse files", err)
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

	runner, err := tflint.NewRunner(cfg, files, annotations, configs, variables...)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to initialize a runner", err)
	}

	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return []*tflint.Runner{}, tflint.NewContextError("Failed to prepare rule checking", err)
	}

	return append(runners, runner), nil
}
