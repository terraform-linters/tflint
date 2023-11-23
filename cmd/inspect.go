package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cli *CLI) inspect(opts Options) int {
	// Respect the "--format" flag until a config is loaded
	cli.formatter.Format = opts.Format

	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	issues := tflint.Issues{}
	changes := map[string][]byte{}

	for _, wd := range workingDirs {
		err := cli.withinChangedDir(wd, func() error {
			filterFiles := []string{}
			for _, pattern := range opts.Filter {
				files, err := filepath.Glob(pattern)
				if err != nil {
					return fmt.Errorf("Failed to parse --filter options; %w", err)
				}
				// Add the raw pattern to return an empty result if it doesn't match any files
				if len(files) == 0 {
					filterFiles = append(filterFiles, pattern)
				}
				filterFiles = append(filterFiles, files...)
			}

			// Join with the working directory to create the fullpath
			for i, file := range filterFiles {
				filterFiles[i] = filepath.Join(wd, file)
			}

			moduleIssues, moduleChanges, err := cli.inspectModule(opts, ".", filterFiles)
			if err != nil {
				return err
			}
			issues = append(issues, moduleIssues...)
			for path, source := range moduleChanges {
				changes[path] = source
			}

			return nil
		})
		if err != nil {
			sources := map[string][]byte{}
			if cli.loader != nil {
				sources = cli.loader.Sources()
			}
			cli.formatter.Print(tflint.Issues{}, err, sources)
			return ExitCodeError
		}
	}

	var force bool
	if opts.Recursive {
		// Respect "--format" and "--force" flags in recursive mode
		cli.formatter.Format = opts.Format
		if opts.Force != nil {
			force = *opts.Force
		}
	} else {
		cli.formatter.Format = cli.config.Format
		force = cli.config.Force
	}

	cli.formatter.Fix = opts.Fix
	cli.formatter.Print(issues, nil, cli.sources)

	if opts.Fix {
		if err := writeChanges(changes); err != nil {
			cli.formatter.Print(tflint.Issues{}, err, cli.sources)
			return ExitCodeError
		}
	}

	if len(issues) > 0 && !force && exceedsMinimumFailure(issues, opts.MinimumFailureSeverity) {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) inspectModule(opts Options, dir string, filterFiles []string) (tflint.Issues, map[string][]byte, error) {
	issues := tflint.Issues{}
	changes := map[string][]byte{}
	var err error

	// Setup config
	cli.config, err = tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
	if err != nil {
		return issues, changes, fmt.Errorf("Failed to load TFLint config; %w", err)
	}
	cli.config.Merge(opts.toConfig())

	// Setup loader
	cli.loader, err = terraform.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, cli.originalWorkingDir)
	if err != nil {
		return issues, changes, fmt.Errorf("Failed to prepare loading; %w", err)
	}
	if opts.Recursive && !cli.loader.IsConfigDir(dir) {
		// Ignore non-module directories in recursive mode
		return issues, changes, nil
	}

	// Setup runners
	runners, err := cli.setupRunners(opts, dir)
	if err != nil {
		return issues, changes, err
	}
	rootRunner := runners[len(runners)-1]

	// Launch plugin processes
	rulesetPlugin, err := launchPlugins(cli.config, opts.Fix)
	if rulesetPlugin != nil {
		defer rulesetPlugin.Clean()
	}
	if err != nil {
		return issues, changes, err
	}

	// Check preconditions
	sdkVersions := map[string]*version.Version{}
	for name, ruleset := range rulesetPlugin.RuleSets {
		sdkVersion, err := ruleset.SDKVersion()
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
				// SDKVersion endpoint is available in tflint-plugin-sdk v0.14+.
				return issues, changes, fmt.Errorf(`Plugin "%s" SDK version is incompatible. Compatible versions: %s`, name, plugin.SDKVersionConstraints)
			} else {
				return issues, changes, fmt.Errorf(`Failed to get plugin "%s" SDK version; %w`, name, err)
			}
		}
		if !plugin.SDKVersionConstraints.Check(sdkVersion) {
			return issues, changes, fmt.Errorf(`Plugin "%s" SDK version (%s) is incompatible. Compatible versions: %s`, name, sdkVersion, plugin.SDKVersionConstraints)
		}
		sdkVersions[name] = sdkVersion
	}

	// Run inspection
	//
	// Repeat an inspection until there are no more changes or the limit is reached,
	// in case an autofix introduces new issues.
	for loop := 1; ; loop++ {
		if loop > 10 {
			return issues, changes, fmt.Errorf(`Reached the limit of autofix attempts, and the changes made by the autofix will not be applied. This may be due to the following reasons:

1. The autofix is making changes that do not fix the issue.
2. The autofix is continuing to introduce new issues.

By setting TFLINT_LOG=trace, you can confirm the changes made by the autofix and start troubleshooting.`)
		}

		for name, ruleset := range rulesetPlugin.RuleSets {
			for _, runner := range runners {
				err = ruleset.Check(plugin.NewGRPCServer(runner, rootRunner, cli.loader.Files(), sdkVersions[name]))
				if err != nil {
					return issues, changes, fmt.Errorf("Failed to check ruleset; %w", err)
				}
			}
		}

		changesInAttempt := map[string][]byte{}
		for _, runner := range runners {
			for _, issue := range runner.LookupIssues(filterFiles...) {
				// On the second attempt, only fixable issues are appended to avoid duplicates.
				if loop == 1 || issue.Fixable {
					issues = append(issues, issue)
				}
			}
			runner.Issues = tflint.Issues{}

			for path, source := range runner.LookupChanges(filterFiles...) {
				changesInAttempt[path] = source
				changes[path] = source
			}
			runner.ClearChanges()
		}

		if !opts.Fix || len(changesInAttempt) == 0 {
			break
		}
	}

	// Set module sources to CLI
	for path, source := range cli.loader.Sources() {
		cli.sources[path] = source
	}

	return issues, changes, nil
}

func (cli *CLI) setupRunners(opts Options, dir string) ([]*tflint.Runner, error) {
	configs, diags := cli.loader.LoadConfig(dir, cli.config.CallModuleType)
	if diags.HasErrors() {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load configurations; %w", diags)
	}

	files, diags := cli.loader.LoadConfigDirFiles(dir)
	if diags.HasErrors() {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load configurations; %w", diags)
	}
	annotations := map[string]tflint.Annotations{}
	for path, file := range files {
		if !strings.HasSuffix(path, ".tf") {
			continue
		}
		ants, lexDiags := tflint.NewAnnotations(path, file)
		diags = diags.Extend(lexDiags)
		annotations[path] = ants
	}
	if diags.HasErrors() {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load configurations; %w", diags)
	}

	variables, diags := cli.loader.LoadValuesFiles(dir, cli.config.Varfiles...)
	if diags.HasErrors() {
		return []*tflint.Runner{}, fmt.Errorf("Failed to load values files; %w", diags)
	}
	cliVars, diags := terraform.ParseVariableValues(cli.config.Variables, configs.Module.Variables)
	if diags.HasErrors() {
		return []*tflint.Runner{}, fmt.Errorf("Failed to parse variables; %w", diags)
	}
	variables = append(variables, cliVars)

	runner, err := tflint.NewRunner(cli.originalWorkingDir, cli.config, annotations, configs, variables...)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to initialize a runner; %w", err)
	}

	runners, err := tflint.NewModuleRunners(runner)
	if err != nil {
		return []*tflint.Runner{}, fmt.Errorf("Failed to prepare rule checking; %w", err)
	}

	return append(runners, runner), nil
}

func launchPlugins(config *tflint.Config, fix bool) (*plugin.Plugin, error) {
	// Lookup plugins
	rulesetPlugin, err := plugin.Discovery(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize plugins; %w", err)
	}

	rulesets := []tflint.RuleSet{}
	pluginConf := config.ToPluginConfig()
	pluginConf.Fix = fix

	// Check version constraints and apply a config to plugins
	for name, ruleset := range rulesetPlugin.RuleSets {
		constraints, err := ruleset.VersionConstraints()
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
				// VersionConstraints endpoint is available in tflint-plugin-sdk v0.14+.
				return rulesetPlugin, fmt.Errorf(`Plugin "%s" SDK version is incompatible. Compatible versions: %s`, name, plugin.SDKVersionConstraints)
			} else {
				return rulesetPlugin, fmt.Errorf("Failed to get TFLint version constraints to `%s` plugin; %w", name, err)
			}
		}
		if !constraints.Check(tflint.Version) {
			return rulesetPlugin, fmt.Errorf("Failed to satisfy version constraints; tflint-ruleset-%s requires %s, but TFLint version is %s", name, constraints, tflint.Version)
		}

		if err := ruleset.ApplyGlobalConfig(pluginConf); err != nil {
			return rulesetPlugin, fmt.Errorf("Failed to apply global config to `%s` plugin; %w", name, err)
		}
		configSchema, err := ruleset.ConfigSchema()
		if err != nil {
			return rulesetPlugin, fmt.Errorf("Failed to fetch config schema from `%s` plugin; %w", name, err)
		}
		content := &hclext.BodyContent{}
		if plugin, exists := config.Plugins[name]; exists {
			var diags hcl.Diagnostics
			content, diags = plugin.Content(configSchema)
			if diags.HasErrors() {
				return rulesetPlugin, fmt.Errorf("Failed to parse `%s` plugin config; %w", name, diags)
			}
		}
		err = ruleset.ApplyConfig(content, config.Sources())
		if err != nil {
			return rulesetPlugin, fmt.Errorf("Failed to apply config to `%s` plugin; %w", name, err)
		}

		rulesets = append(rulesets, ruleset)
	}

	// Validate config for plugins
	if err := config.ValidateRules(rulesets...); err != nil {
		return rulesetPlugin, fmt.Errorf("Failed to check rule config; %w", err)
	}

	return rulesetPlugin, nil
}

func writeChanges(changes map[string][]byte) error {
	fs := afero.NewOsFs()
	for path, source := range changes {
		f, err := fs.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("Failed to apply autofixes; failed to open %s: %w", path, err)
		}

		n, err := f.Write(source)
		if err == nil && n < len(source) {
			err = io.ErrShortWrite
		}
		if err1 := f.Close(); err == nil {
			err = err1
		}
		if err != nil {
			return fmt.Errorf("Failed to apply autofixes; failed to write source code to %s: %w", path, err)
		}
	}
	return nil
}

// Checks if the given issues contain severities above or equal to the given minimum failure opt. Defaults to true if an error occurs
func exceedsMinimumFailure(issues tflint.Issues, minimumFailureOpt string) bool {
	if minimumFailureOpt != "" {
		minSeverity, err := tflint.NewSeverity(minimumFailureOpt)
		if err != nil {
			return true
		}

		minSeverityInt32, err := tflint.SeverityToInt32(minSeverity)
		if err != nil {
			return true
		}

		for _, i := range issues {
			ruleSeverityInt32, err := tflint.SeverityToInt32(i.Rule.Severity())
			if err != nil {
				return true
			}
			if ruleSeverityInt32 >= minSeverityInt32 {
				return true
			}
		}
		return false
	}
	return true
}
