package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) printVersion(opts Options) int {
	fmt.Fprintf(cli.outStream, "TFLint version %s\n", tflint.Version)

	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	if opts.Recursive {
		fmt.Fprint(cli.outStream, "\n")
	}

	for _, wd := range workingDirs {
		err := cli.withinChangedDir(wd, func() error {
			if opts.Recursive {
				fmt.Fprint(cli.outStream, "====================================================\n")
				fmt.Fprintf(cli.outStream, "working directory: %s\n\n", wd)
			}

			versions := getPluginVersions(opts)

			for _, version := range versions {
				fmt.Fprint(cli.outStream, version)
			}
			if len(versions) == 0 && opts.Recursive {
				fmt.Fprint(cli.outStream, "No plugins\n")
			}
			return nil
		})
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, err, map[string][]byte{})
		}
	}

	return ExitCodeOK
}

func getPluginVersions(opts Options) []string {
	// Load configuration files to print plugin versions
	cfg, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
	if err != nil {
		log.Printf("[ERROR] Failed to load TFLint config: %s", err)
		return []string{}
	}
	cfg.Merge(opts.toConfig())

	rulesetPlugin, err := plugin.Discovery(cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to initialize plugins: %s", err)
		return []string{}
	}
	defer rulesetPlugin.Clean()

	versions := []string{}
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

		versions = append(versions, fmt.Sprintf("+ ruleset.%s (%s)\n", name, version))
	}

	return versions
}
