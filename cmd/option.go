package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
)

// Options is an option specified by arguments.
type Options struct {
	Version                bool     `short:"v" long:"version" description:"Print TFLint version"`
	Init                   bool     `long:"init" description:"Install plugins"`
	Langserver             bool     `long:"langserver" description:"Start language server"`
	Format                 string   `short:"f" long:"format" description:"Output format" choice:"default" choice:"json" choice:"checkstyle" choice:"junit" choice:"compact" choice:"sarif"`
	Config                 string   `short:"c" long:"config" description:"Config file name (default: .tflint.hcl)" value-name:"FILE"`
	IgnoreModules          []string `long:"ignore-module" description:"Ignore module sources" value-name:"SOURCE"`
	EnableRules            []string `long:"enable-rule" description:"Enable rules from the command line" value-name:"RULE_NAME"`
	DisableRules           []string `long:"disable-rule" description:"Disable rules from the command line" value-name:"RULE_NAME"`
	Only                   []string `long:"only" description:"Enable only this rule, disabling all other defaults. Can be specified multiple times" value-name:"RULE_NAME"`
	EnablePlugins          []string `long:"enable-plugin" description:"Enable plugins from the command line" value-name:"PLUGIN_NAME"`
	Varfiles               []string `long:"var-file" description:"Terraform variable file name" value-name:"FILE"`
	Variables              []string `long:"var" description:"Set a Terraform variable" value-name:"'foo=bar'"`
	Module                 *bool    `long:"module" description:"Enable module inspection" hidden:"true"`
	NoModule               *bool    `long:"no-module" description:"Disable module inspection" hidden:"true"`
	CallModuleType         *string  `long:"call-module-type" description:"Types of module to call (default: local)" choice:"all" choice:"local" choice:"none"`
	Chdir                  string   `long:"chdir" description:"Switch to a different working directory before executing the command" value-name:"DIR"`
	Recursive              bool     `long:"recursive" description:"Run command in each directory recursively"`
	Filter                 []string `long:"filter" description:"Filter issues by file names or globs" value-name:"FILE"`
	Force                  *bool    `long:"force" description:"Return zero exit status even if issues found"`
	MinimumFailureSeverity string   `long:"minimum-failure-severity" description:"Sets minimum severity level for exiting with a non-zero error code" choice:"error" choice:"warning" choice:"notice"`
	Color                  bool     `long:"color" description:"Enable colorized output"`
	NoColor                bool     `long:"no-color" description:"Disable colorized output"`
	Fix                    bool     `long:"fix" description:"Fix issues automatically"`
	NoParallelRunners      bool     `long:"no-parallel-runners" description:"Disable per-runner parallelism"`
	MaxWorkers             *int     `long:"max-workers" description:"Set maximum number of workers in recursive inspection (default: number of CPUs)" value-name:"N"`
	ActAsBundledPlugin     bool     `long:"act-as-bundled-plugin" hidden:"true"`
	ActAsWorker            bool     `long:"act-as-worker" hidden:"true"`
}

func (opts *Options) toConfig() *tflint.Config {
	ignoreModules := map[string]bool{}
	for _, module := range opts.IgnoreModules {
		// For the backward compatibility, allow specifying like "source1,source2" style
		for _, m := range strings.Split(module, ",") {
			ignoreModules[m] = true
		}
	}

	varfiles := []string{}
	for _, vf := range opts.Varfiles {
		// For the backward compatibility, allow specifying like "varfile1,varfile2" style
		varfiles = append(varfiles, strings.Split(vf, ",")...)
	}
	if opts.Variables == nil {
		opts.Variables = []string{}
	}

	callModuleType := terraform.CallLocalModule
	callModuleTypeSet := false
	// --call-module-type takes precedence over --module/--no-module. This is for backward compatibility.
	if opts.Module != nil {
		fmt.Fprintln(os.Stderr, "WARNING: --module is deprecated. Use --call-module-type=all instead.")
		callModuleType = terraform.CallAllModule
		callModuleTypeSet = true
	}
	if opts.NoModule != nil {
		fmt.Fprintln(os.Stderr, "WARNING: --no-module is deprecated. Use --call-module-type=none instead.")
		callModuleType = terraform.CallNoModule
		callModuleTypeSet = true
	}
	if opts.CallModuleType != nil {
		var err error
		callModuleType, err = terraform.AsCallModuleType(*opts.CallModuleType)
		if err != nil {
			// This should never happen because the option is already validated by go-flags
			panic(err)
		}
		callModuleTypeSet = true
	}

	var force, forceSet bool
	if opts.Force != nil {
		force = *opts.Force
		forceSet = true
	}

	log.Printf("[DEBUG] CLI Options")
	log.Printf("[DEBUG]   CallModuleType: %s", callModuleType)
	log.Printf("[DEBUG]   Force: %t", force)
	log.Printf("[DEBUG]   Format: %s", opts.Format)
	log.Printf("[DEBUG]   Varfiles: %s", strings.Join(opts.Varfiles, ", "))
	log.Printf("[DEBUG]   Variables: %s", strings.Join(opts.Variables, ", "))
	log.Printf("[DEBUG]   EnableRules: %s", strings.Join(opts.EnableRules, ", "))
	log.Printf("[DEBUG]   DisableRules: %s", strings.Join(opts.DisableRules, ", "))
	log.Printf("[DEBUG]   Only: %s", strings.Join(opts.Only, ", "))
	log.Printf("[DEBUG]   EnablePlugins: %s", strings.Join(opts.EnablePlugins, ", "))
	log.Printf("[DEBUG]   IgnoreModules:")
	for name, ignore := range ignoreModules {
		log.Printf("[DEBUG]     %s: %t", name, ignore)
	}

	rules := map[string]*tflint.RuleConfig{}
	for _, rule := range append(opts.Only, opts.EnableRules...) {
		rules[rule] = &tflint.RuleConfig{
			Name:    rule,
			Enabled: true,
			Body:    nil,
		}
	}
	for _, rule := range opts.DisableRules {
		rules[rule] = &tflint.RuleConfig{
			Name:    rule,
			Enabled: false,
			Body:    nil,
		}
	}

	if len(opts.Only) > 0 && (len(opts.EnableRules) > 0 || len(opts.DisableRules) > 0) {
		log.Printf("[WARN] Usage of --only will ignore other rules provided by --enable-rule or --disable-rule")
	}

	plugins := map[string]*tflint.PluginConfig{}
	for _, plugin := range opts.EnablePlugins {
		plugins[plugin] = &tflint.PluginConfig{
			Name:    plugin,
			Enabled: true,
			Body:    nil,
		}
	}

	return &tflint.Config{
		CallModuleType:    callModuleType,
		CallModuleTypeSet: callModuleTypeSet,

		Force:    force,
		ForceSet: forceSet,

		Format:    opts.Format,
		FormatSet: opts.Format != "",

		DisabledByDefault:    len(opts.Only) > 0,
		DisabledByDefaultSet: len(opts.Only) > 0,

		Varfiles:      varfiles,
		Variables:     opts.Variables,
		Only:          opts.Only,
		IgnoreModules: ignoreModules,
		Rules:         rules,
		Plugins:       plugins,
	}
}

// Return commands to be executed by worker processes in recursive inspection.
// All possible CLI flags are delegated, but some flags are ignored because
// the coordinator process that starts the workers is responsible.
func (opts *Options) toWorkerCommands(workingDir string) []string {
	commands := []string{
		"--act-as-worker",
		"--chdir=" + workingDir,
		"--force", // Exit status is always ignored
	}

	// opts.Version, opts.Init, and opts.Langserver are not supported

	// opt.Format is ignored because workers always output serialized issues

	if opts.Config != "" {
		commands = append(commands, fmt.Sprintf("--config=%s", opts.Config))
	}
	for _, ignoreModule := range opts.IgnoreModules {
		commands = append(commands, fmt.Sprintf("--ignore-module=%s", ignoreModule))
	}
	for _, rule := range opts.EnableRules {
		commands = append(commands, fmt.Sprintf("--enable-rule=%s", rule))
	}
	for _, rule := range opts.DisableRules {
		commands = append(commands, fmt.Sprintf("--disable-rule=%s", rule))
	}
	for _, rule := range opts.Only {
		commands = append(commands, fmt.Sprintf("--only=%s", rule))
	}
	for _, plugin := range opts.EnablePlugins {
		commands = append(commands, fmt.Sprintf("--enable-plugin=%s", plugin))
	}
	for _, varfile := range opts.Varfiles {
		commands = append(commands, fmt.Sprintf("--var-file=%s", varfile))
	}
	for _, variable := range opts.Variables {
		commands = append(commands, fmt.Sprintf("--var=%s", variable))
	}
	if opts.Module != nil && *opts.Module {
		commands = append(commands, "--module")
	}
	if opts.NoModule != nil && *opts.NoModule {
		commands = append(commands, "--no-module")
	}
	if opts.CallModuleType != nil {
		commands = append(commands, fmt.Sprintf("--call-module-type=%s", *opts.CallModuleType))
	}

	// opts.Chdir should be ignored because it is given by the coordinator

	// opts.Recursive is not supported

	for _, filter := range opts.Filter {
		commands = append(commands, fmt.Sprintf("--filter=%s", filter))
	}

	// opts.Force and opts.MinimumFailureSeverity are ignored because exit status is controlled by the coordinator

	// opts.Color and opts.NoColor are ignored because the coordinator is responsible for colorized output

	if opts.Fix {
		commands = append(commands, "--fix")
	}
	if opts.NoParallelRunners {
		commands = append(commands, "--no-parallel-runners")
	}

	// opts.MaxWorkers is ignored because the coordinator is responsible for parallelism

	// opts.ActAsBundledPlugin and opts.ActAsWorker are not supported

	return commands
}
