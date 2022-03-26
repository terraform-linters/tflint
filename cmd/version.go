package cmd

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) printVersion(opts Options) int {
	fmt.Fprintf(cli.outStream, "TFLint version %s\n", tflint.Version)

	// Load configuration files to print plugin versions
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		log.Printf("[ERROR] Failed to load TFLint config: %s", err)
		return ExitCodeOK
	}
	if len(opts.Only) > 0 {
		for _, rule := range cfg.Rules {
			rule.Enabled = false
		}
	}
	cfg = cfg.Merge(opts.toConfig())

	rulesetPlugin, err := plugin.Discovery(cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to initialize plugins: %s", err)
		return ExitCodeOK
	}
	defer rulesetPlugin.Clean()

	for _, ruleset := range rulesetPlugin.RuleSets {
		name, err := ruleset.RuleSetName()
		if err != nil {
			log.Printf("[ERROR] Failed to get ruleset name: %s", err)
			continue
		}
		version, err := ruleset.RuleSetVersion()
		if err != nil {
			log.Printf("[ERROR] Failed to get ruleset version: %s", err)
			continue
		}

		fmt.Fprintf(cli.outStream, "+ ruleset.%s (%s)\n", name, version)
	}

	return ExitCodeOK
}
