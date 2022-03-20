package cmd

import (
	"log"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// Options is an option specified by arguments.
type Options struct {
	Version       bool     `short:"v" long:"version" description:"Print TFLint version"`
	Init          bool     `long:"init" description:"Install plugins"`
	Langserver    bool     `long:"langserver" description:"Start language server"`
	Format        string   `short:"f" long:"format" description:"Output format" choice:"default" choice:"json" choice:"checkstyle" choice:"junit" choice:"compact" choice:"sarif" default:"default"`
	Config        string   `short:"c" long:"config" description:"Config file name" value-name:"FILE" default:".tflint.hcl"`
	IgnoreModules []string `long:"ignore-module" description:"Ignore module sources" value-name:"SOURCE"`
	EnableRules   []string `long:"enable-rule" description:"Enable rules from the command line" value-name:"RULE_NAME"`
	DisableRules  []string `long:"disable-rule" description:"Disable rules from the command line" value-name:"RULE_NAME"`
	Only          []string `long:"only" description:"Enable only this rule, disabling all other defaults. Can be specified multiple times" value-name:"RULE_NAME"`
	EnablePlugins []string `long:"enable-plugin" description:"Enable plugins from the command line" value-name:"PLUGIN_NAME"`
	Varfiles      []string `long:"var-file" description:"Terraform variable file name" value-name:"FILE"`
	Variables     []string `long:"var" description:"Set a Terraform variable" value-name:"'foo=bar'"`
	Module        bool     `long:"module" description:"Inspect modules"`
	Force         bool     `long:"force" description:"Return zero exit status even if issues found"`
	NoColor       bool     `long:"no-color" description:"Disable colorized output"`
	LogLevel      string   `long:"loglevel" description:"Change the loglevel" choice:"trace" choice:"debug" choice:"info" choice:"warn" choice:"error"`
}

func (opts *Options) toConfig() *tflint.Config {
	ignoreModules := map[string]bool{}
	for _, module := range opts.IgnoreModules {
		// For the backward compatibility, allow specifying like `source1,source2` style
		for _, m := range strings.Split(module, ",") {
			ignoreModules[m] = true
		}
	}

	varfiles := []string{}
	for _, vf := range opts.Varfiles {
		// For the backward compatibility, allow specifying like `varfile1,varfile2` style
		varfiles = append(varfiles, strings.Split(vf, ",")...)
	}
	if opts.Variables == nil {
		opts.Variables = []string{}
	}

	log.Printf("[DEBUG] CLI Options")
	log.Printf("[DEBUG]   Module: %t", opts.Module)
	log.Printf("[DEBUG]   Force: %t", opts.Force)
	log.Printf("[DEBUG]   IgnoreModules: %#v", ignoreModules)
	log.Printf("[DEBUG]   EnableRules: %#v", opts.EnableRules)
	log.Printf("[DEBUG]   DisableRules: %#v", opts.DisableRules)
	log.Printf("[DEBUG]   Only: %#v", opts.Only)
	log.Printf("[DEBUG]   EnablePlugins: %#v", opts.EnablePlugins)
	log.Printf("[DEBUG]   Varfiles: %#v", varfiles)
	log.Printf("[DEBUG]   Variables: %#v", opts.Variables)

	rules := map[string]*tflint.RuleConfig{}
	if len(opts.Only) > 0 {
		for _, rule := range opts.Only {
			rules[rule] = &tflint.RuleConfig{
				Name:    rule,
				Enabled: true,
				Body:    hcl.EmptyBody(),
			}
		}
	} else {
		for _, rule := range opts.EnableRules {
			rules[rule] = &tflint.RuleConfig{
				Name:    rule,
				Enabled: true,
				Body:    hcl.EmptyBody(),
			}
		}
		for _, rule := range opts.DisableRules {
			rules[rule] = &tflint.RuleConfig{
				Name:    rule,
				Enabled: false,
				Body:    hcl.EmptyBody(),
			}
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
		Module:            opts.Module,
		Force:             opts.Force,
		IgnoreModules:     ignoreModules,
		Varfiles:          varfiles,
		Variables:         opts.Variables,
		DisabledByDefault: len(opts.Only) > 0,
		Rules:             rules,
		Plugins:           plugins,
	}
}
