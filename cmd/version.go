package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/afero"
	tfplugin "github.com/terraform-linters/tflint/plugin"
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

	cli.loader, err = tflint.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to prepare loading: %s", err)
		return ExitCodeOK
	}

	runners, appErr := cli.setupRunners(opts, cfg, ".")
	if appErr != nil {
		log.Printf("[ERROR] Failed to setup runners: %s", appErr)
		return ExitCodeOK
	}
	rootRunner := runners[len(runners)-1]

	// AWS plugin is automatically enabled from your provider requirements, even if the plugin isn't explicitly enabled.
	if _, exists := cfg.Plugins["aws"]; !exists {
		reqs, diags := rootRunner.TFConfig.ProviderRequirements()
		if diags.HasErrors() {
			log.Printf("[ERROR] Failed to get Terraform provider requirements: %s", diags)
			return ExitCodeOK
		}
		for addr := range reqs {
			if addr.Type == "aws" {
				log.Print("[INFO] AWS provider requirements found. Enable the plugin `aws` automatically")
				cfg.Plugins["aws"] = &tflint.PluginConfig{
					Name:    "aws",
					Enabled: true,
				}
			}
		}
	}

	plugin, err := tfplugin.Discovery(cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to initialize plugins: %s", err)
		return ExitCodeOK
	}
	defer plugin.Clean()

	for _, ruleset := range plugin.RuleSets {
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
