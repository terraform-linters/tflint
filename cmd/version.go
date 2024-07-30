package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/go-github/v35/github"
	"github.com/hashicorp/go-version"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) printVersion(opts Options) int {
	fmt.Fprintf(cli.outStream, "TFLint version %s\n", tflint.Version)

	cli.printLatestReleaseVersion()

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

// Checks GitHub releases and prints new version, if current version is outdated.
// requires GitHub releases to follow semver.
func (cli *CLI) printLatestReleaseVersion() {
	latest, err := getLatestVersion()
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to check updates; %w", err), map[string][]byte{})
	}
	latestVersion, err := version.NewSemver(*latest.Name)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to parse version; %w", err), map[string][]byte{})
	}
	compare := tflint.Version.Compare(latestVersion)
	if compare < 0 {
		fmt.Fprintf(cli.outStream, "New version available: %s\n", *latest.HTMLURL)
	}
}

func getLatestVersion() (*github.RepositoryRelease, error) {
	ghClient := github.NewClient(nil)
	releases, _, err := ghClient.Repositories.ListReleases(context.Background(),
		"terraform-linters", "tflint", &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	// GitHub sorts releases results. Select first non-prerelease version and return it.
	for i := range releases {
		release := releases[i]
		if !*release.Prerelease {
			return release, nil
		}
	}
	return nil, errors.New("not found")
}
