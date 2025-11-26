package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"slices"
	"time"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/terraform-linters/tflint/versioncheck"
)

const (
	versionCheckTimeout = 3 * time.Second
)

// VersionOutput is the JSON output structure for version command
type VersionOutput struct {
	Version         string          `json:"version"`
	Plugins         []PluginVersion `json:"plugins"`
	UpdateAvailable bool            `json:"update_available"`
	LatestVersion   string          `json:"latest_version,omitempty"`
}

// PluginVersion represents a plugin's name and version
type PluginVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (cli *CLI) printVersion(opts Options) int {
	// Check for updates (unless disabled)
	var updateInfo *versioncheck.UpdateInfo
	if versioncheck.Enabled() {
		ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
		defer cancel()

		info, err := versioncheck.CheckForUpdate(ctx, tflint.Version)
		if err != nil {
			log.Printf("[ERROR] Failed to check for updates: %s", err)
		} else {
			updateInfo = info
		}
	}

	// If JSON format requested, output JSON
	if opts.Format == "json" {
		return cli.printVersionJSON(opts, updateInfo)
	}

	// Print version
	fmt.Fprintf(cli.outStream, "TFLint version %s\n", tflint.Version)

	// Print update notification if available
	if updateInfo != nil && updateInfo.Available {
		fmt.Fprintf(cli.outStream, "\n")
		fmt.Fprintf(cli.outStream, "Your version of TFLint is out of date! The latest version\n")
		fmt.Fprintf(cli.outStream, "is %s. You can update by downloading from https://github.com/terraform-linters/tflint/releases\n", updateInfo.Latest)
	}

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

			plugins := getPluginVersions(opts)

			for _, plugin := range plugins {
				fmt.Fprintf(cli.outStream, "+ %s (%s)\n", plugin.Name, plugin.Version)
			}
			if len(plugins) == 0 && opts.Recursive {
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

func (cli *CLI) printVersionJSON(opts Options, updateInfo *versioncheck.UpdateInfo) int {
	// Build output
	output := VersionOutput{
		Version: tflint.Version.String(),
		Plugins: getPluginVersions(opts),
	}

	if updateInfo != nil {
		output.UpdateAvailable = updateInfo.Available
		if updateInfo.Available {
			output.LatestVersion = updateInfo.Latest
		}
	}

	// Marshal and print JSON
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Failed to marshal JSON: %s", err)
		return ExitCodeError
	}

	fmt.Fprintln(cli.outStream, string(jsonBytes))
	return ExitCodeOK
}

func getPluginVersions(opts Options) []PluginVersion {
	cfg, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
	if err != nil {
		log.Printf("[ERROR] Failed to load TFLint config: %s", err)
		return []PluginVersion{}
	}
	cfg.Merge(opts.toConfig())

	rulesetPlugin, err := plugin.Discovery(cfg)
	if err != nil {
		log.Printf("[ERROR] Failed to initialize plugins: %s", err)
		return []PluginVersion{}
	}
	defer rulesetPlugin.Clean()

	// Sort ruleset names to ensure consistent ordering
	rulesetNames := slices.Sorted(maps.Keys(rulesetPlugin.RuleSets))

	plugins := []PluginVersion{}
	for _, name := range rulesetNames {
		ruleset := rulesetPlugin.RuleSets[name]
		rulesetName, err := ruleset.RuleSetName()
		if err != nil {
			log.Printf("[ERROR] Failed to get ruleset name: %s", err)
			continue
		}
		version, err := ruleset.RuleSetVersion()
		if err != nil {
			log.Printf("[ERROR] Failed to get ruleset version: %s", err)
			continue
		}

		plugins = append(plugins, PluginVersion{
			Name:    fmt.Sprintf("ruleset.%s", rulesetName),
			Version: version,
		})
	}

	return plugins
}
