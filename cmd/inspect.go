package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/internal/workingdir"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) inspect(opts Options) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	issues, changes, err := cli.inspectDirectory(ctx, opts)
	if ctx.Err() != nil {
		return ExitCodeError
	}
	if err != nil {
		sources := map[string][]byte{}
		if cli.loader != nil {
			sources = cli.loader.Sources()
		}
		cli.formatter.Print(tflint.Issues{}, err, sources)
		return ExitCodeError
	}

	if opts.ActAsWorker {
		// When acting as a recursive inspection worker, the formatter is ignored
		// and the serialized issues are output.
		out, err := json.Marshal(issues)
		if err != nil {
			fmt.Fprint(cli.errStream, err)
			return ExitCodeError
		}
		fmt.Fprint(cli.outStream, string(out))
	} else {
		cli.formatter.Print(issues, nil, cli.sources)
	}

	if opts.Fix {
		if err := writeChanges(changes); err != nil {
			cli.formatter.Print(tflint.Issues{}, err, cli.sources)
			return ExitCodeError
		}
	}

	if len(issues) > 0 && !cli.config.Force && exceedsMinimumFailure(issues, opts.MinimumFailureSeverity) {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func (cli *CLI) inspectDirectory(ctx context.Context, opts Options) (tflint.Issues, map[string][]byte, error) {
	wd := workingdir.Resolve(cli.originalWorkingDir, opts.Chdir)
	filterFiles := []string{}
	for _, pattern := range opts.Filter {
		pattern = workingdir.Resolve(wd, pattern)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to parse --filter options; %w", err)
		}
		if len(files) == 0 {
			files = []string{pattern}
		}
		for _, file := range files {
			path, err := filepath.Rel(cli.originalWorkingDir, file)
			if err != nil {
				return nil, nil, err
			}
			filterFiles = append(filterFiles, path)
		}
	}
	return cli.inspectModule(ctx, opts, wd, filterFiles)
}

func (cli *CLI) inspectModule(ctx context.Context, opts Options, wd string, filterFiles []string) (tflint.Issues, map[string][]byte, error) {
	issues := tflint.Issues{}
	changes := map[string][]byte{}
	var err error

	// Setup config
	cli.config, err = tflint.LoadConfig(afero.Afero{Fs: workingdir.New(afero.NewOsFs(), wd)}, opts.Config)
	if err != nil {
		return issues, changes, fmt.Errorf("Failed to load TFLint config; %w", err)
	}
	cli.config.Merge(opts.toConfig())
	// Apply format set in config file
	cli.formatter.Format = cli.config.Format

	// Setup loader
	cli.loader, err = terraform.NewLoaderAt(afero.Afero{Fs: afero.NewOsFs()}, cli.originalWorkingDir, wd)
	if err != nil {
		return issues, changes, fmt.Errorf("Failed to prepare loading; %w", err)
	}
	if opts.ActAsWorker && !cli.loader.IsConfigDir(".") {
		// Ignore non-module directories in worker mode
		return issues, changes, nil
	}

	// Setup runners
	rootRunner, moduleRunners, err := tflint.BuildRunners(cli.loader, cli.config, cli.originalWorkingDir, ".")
	if err != nil {
		return issues, changes, err
	}

	// Launch plugin processes
	rulesetPlugin, err := launchPlugins(cli.config, opts.Fix, wd)
	if rulesetPlugin != nil {
		defer rulesetPlugin.Clean()
		stop := context.AfterFunc(ctx, rulesetPlugin.Clean)
		defer stop()
	}
	if err != nil {
		return issues, changes, err
	}

	// Validate and collect plugin versions
	sdkVersions, err := plugin.ValidatePluginVersions(rulesetPlugin, cli.config.IsJSONConfig())
	if err != nil {
		return issues, changes, err
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
			if err := ruleset.Check(plugin.NewGRPCServer(rootRunner, rootRunner, cli.loader.Files(), sdkVersions[name])); err != nil {
				return issues, changes, fmt.Errorf("Failed to check ruleset; %w", err)
			}
			// Run checks for module calls are performed in parallel.
			// The rootRunner is shared between goroutines but read-only, so this is goroutine-safe.
			// Note that checks against the rootRunner are not parallelized, as autofix may cause the module to be rebuilt.
			ch := make(chan error, len(moduleRunners))
			for _, runner := range moduleRunners {
				if opts.NoParallelRunners {
					ch <- ruleset.Check(plugin.NewGRPCServer(runner, rootRunner, cli.loader.Files(), sdkVersions[name]))
				} else {
					go func(runner *tflint.Runner) {
						ch <- ruleset.Check(plugin.NewGRPCServer(runner, rootRunner, cli.loader.Files(), sdkVersions[name]))
					}(runner)
				}
			}
			var checkErr error
			for range moduleRunners {
				err = <-ch
				if err != nil {
					checkErr = err
				}
			}
			close(ch)
			if checkErr != nil {
				return issues, changes, fmt.Errorf("Failed to check ruleset; %w", checkErr)
			}
		}

		changesInAttempt := map[string][]byte{}
		for _, runner := range append(moduleRunners, rootRunner) {
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
	maps.Copy(cli.sources, cli.loader.Sources())

	return issues, changes, nil
}

func launchPlugins(config *tflint.Config, fix bool, dir string) (*plugin.Plugin, error) {
	// Lookup plugins
	rulesetPlugin, err := plugin.DiscoveryAt(config, dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize plugins; %w", err)
	}

	rulesets := []tflint.RuleSet{}
	pluginConf := config.ToPluginConfig()
	pluginConf.Fix = fix

	// Apply config to plugins
	for name, ruleset := range rulesetPlugin.RuleSets {
		// Check TFLint version constraints before applying config
		constraints, err := ruleset.VersionConstraints()
		if err != nil {
			if plugin.IsVersionConstraintsUnimplemented(err) {
				// VersionConstraints endpoint is available in tflint-plugin-sdk v0.14+.
				// Plugin is too old
				return rulesetPlugin, fmt.Errorf(`Plugin "%s" SDK version is incompatible. Compatible versions: %s`, name, plugin.DefaultSDKVersionConstraints)
			}
			return rulesetPlugin, fmt.Errorf(`Failed to get TFLint version constraints to "%s" plugin; %w`, name, err)
		}
		if err := plugin.CheckTFLintVersionConstraints(name, constraints); err != nil {
			return rulesetPlugin, err
		}

		if err := ruleset.ApplyGlobalConfig(pluginConf); err != nil {
			return rulesetPlugin, fmt.Errorf(`Failed to apply global config to "%s" plugin; %w`, name, err)
		}
		configSchema, err := ruleset.ConfigSchema()
		if err != nil {
			return rulesetPlugin, fmt.Errorf(`Failed to fetch config schema from "%s" plugin; %w`, name, err)
		}
		content := &hclext.BodyContent{}
		if plugin, exists := config.Plugins[name]; exists {
			var diags hcl.Diagnostics
			content, diags = plugin.Content(configSchema)
			if diags.HasErrors() {
				return rulesetPlugin, fmt.Errorf(`Failed to parse "%s" plugin config; %w`, name, diags)
			}
		}
		err = ruleset.ApplyConfig(content, config.Sources())
		if err != nil {
			return rulesetPlugin, fmt.Errorf(`Failed to apply config to "%s" plugin; %w`, name, err)
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
