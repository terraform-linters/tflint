package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cli *CLI) inspect(opts Options, args []string) int {
	// Respect the "--format" flag until a config is loaded
	cli.formatter.Format = opts.Format

	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	issues := tflint.Issues{}

	for _, wd := range workingDirs {
		err := cli.withinChangedDir(wd, func() error {
			// Parse directory/file arguments after changing the working directory
			targetDir, filterFiles, err := processArgs(args[1:])
			if err != nil {
				return fmt.Errorf("Failed to parse CLI arguments; %w", err)
			}

			if opts.Chdir != "" && targetDir != "." {
				return fmt.Errorf("Cannot use --chdir and directory argument at the same time")
			}
			if opts.Recursive && (targetDir != "." || len(filterFiles) > 0) {
				return fmt.Errorf("Cannot use --recursive and arguments at the same time")
			}

			// Join with the working directory to create the fullpath
			for i, file := range filterFiles {
				filterFiles[i] = filepath.Join(wd, file)
			}
			moduleIssues, err := cli.inspectModule(opts, targetDir, filterFiles)
			if err != nil {
				return err
			}
			issues = append(issues, moduleIssues...)
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
		force = opts.Force
	} else {
		cli.formatter.Format = cli.config.Format
		force = cli.config.Force
	}

	cli.formatter.Print(issues, nil, cli.sources)

	if len(issues) > 0 && !force {
		return ExitCodeIssuesFound
	}

	return ExitCodeOK
}

func processArgs(args []string) (string, []string, error) {
	if len(args) == 0 {
		return ".", []string{}, nil
	}

	var dir string
	filterFiles := []string{}

	for _, file := range args {
		fileInfo, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				return dir, filterFiles, fmt.Errorf("Failed to load `%s`: File not found", file)
			}
			return dir, filterFiles, fmt.Errorf("Failed to load `%s`: %s", file, err)
		}

		if fileInfo.IsDir() {
			dir = file
			if len(args) != 1 {
				return dir, filterFiles, fmt.Errorf("Failed to load `%s`: Multiple arguments are not allowed when passing a directory", file)
			}
			return dir, filterFiles, nil
		}

		if !strings.HasSuffix(file, ".tf") && !strings.HasSuffix(file, ".tf.json") {
			return dir, filterFiles, fmt.Errorf("Failed to load `%s`: File is not a target of Terraform", file)
		}

		fileDir := filepath.Dir(file)
		if dir == "" {
			dir = fileDir
			filterFiles = append(filterFiles, file)
		} else if fileDir == dir {
			filterFiles = append(filterFiles, file)
		} else {
			return dir, filterFiles, fmt.Errorf("Failed to load `%s`: Multiple files in different directories are not allowed", file)
		}
	}

	return dir, filterFiles, nil
}

func (cli *CLI) inspectModule(opts Options, dir string, filterFiles []string) (tflint.Issues, error) {
	issues := tflint.Issues{}
	var err error

	// Setup config
	cli.config, err = tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
	if err != nil {
		return tflint.Issues{}, fmt.Errorf("Failed to load TFLint config; %w", err)
	}
	// tflint-plugin-sdk v0.13+ doesn't need to disable rules config when enabling the only option.
	// This is for the backward compatibility.
	if len(opts.Only) > 0 {
		for _, rule := range cli.config.Rules {
			rule.Enabled = false
		}
	}
	cli.config.Merge(opts.toConfig())

	// Setup loader
	cli.loader, err = terraform.NewLoader(afero.Afero{Fs: afero.NewOsFs()}, cli.originalWorkingDir)
	if err != nil {
		return tflint.Issues{}, fmt.Errorf("Failed to prepare loading; %w", err)
	}
	if opts.Recursive && !cli.loader.IsConfigDir(dir) {
		// Ignore non-module directories in recursive mode
		return tflint.Issues{}, nil
	}

	// Setup runners
	runners, err := cli.setupRunners(opts, dir)
	if err != nil {
		return tflint.Issues{}, err
	}
	rootRunner := runners[len(runners)-1]

	// Launch plugin processes
	rulesetPlugin, err := launchPlugins(cli.config)
	if rulesetPlugin != nil {
		defer rulesetPlugin.Clean()
	}
	if err != nil {
		return tflint.Issues{}, err
	}

	// Run inspection
	for _, ruleset := range rulesetPlugin.RuleSets {
		for _, runner := range runners {
			err = ruleset.Check(plugin.NewGRPCServer(runner, rootRunner, cli.loader.Files()))
			if err != nil {
				return tflint.Issues{}, fmt.Errorf("Failed to check ruleset; %w", err)
			}
		}
	}

	for _, runner := range runners {
		issues = append(issues, runner.LookupIssues(filterFiles...)...)
	}
	// Set module sources to CLI
	for path, source := range cli.loader.Sources() {
		cli.sources[path] = source
	}

	return issues, nil
}

func (cli *CLI) setupRunners(opts Options, dir string) ([]*tflint.Runner, error) {
	configs, diags := cli.loader.LoadConfig(dir, cli.config.Module)
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

func launchPlugins(config *tflint.Config) (*plugin.Plugin, error) {
	// Lookup plugins
	rulesetPlugin, err := plugin.Discovery(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize plugins; %w", err)
	}

	rulesets := []tflint.RuleSet{}
	pluginConf := config.ToPluginConfig()

	// Check version constraints and apply a config to plugins
	for name, ruleset := range rulesetPlugin.RuleSets {
		constraints, err := ruleset.VersionConstraints()
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
				// VersionConstraints endpoint is available in tflint-plugin-sdk v0.14+.
				// Skip verification if not available.
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
