package cmd

import (
	"log"
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
	ActAsBundledPlugin     bool     `long:"act-as-bundled-plugin" hidden:"true"`
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
		callModuleType = terraform.CallAllModule
		callModuleTypeSet = true
	}
	if opts.NoModule != nil {
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
