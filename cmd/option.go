package cmd

import (
	"log"
	"strings"

	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/tflint"
)

// Options is an option specified by arguments.
type Options struct {
	Version      bool     `short:"v" long:"version" description:"Print TFLint version"`
	Langserver   bool     `long:"langserver" description:"Start language server"`
	Format       string   `short:"f" long:"format" description:"Output format" choice:"default" choice:"json" choice:"checkstyle" default:"default"`
	Config       string   `short:"c" long:"config" description:"Config file name" value-name:"FILE" default:".tflint.hcl"`
	IgnoreModule string   `long:"ignore-module" description:"Ignore module sources" value-name:"SOURCE1,SOURCE2..."`
	EnableRules  []string `long:"enable-rule" description:"Enable rules from the command line" value-name:"RULE_NAME"`
	DisableRules []string `long:"disable-rule" description:"Disable rules from the command line" value-name:"RULE_NAME"`
	Varfile      string   `long:"var-file" description:"Terraform variable file names" value-name:"FILE1,FILE2..."`
	Variables    []string `long:"var" description:"Set a Terraform variable" value-name:"'foo=bar'"`
	Module       bool     `long:"module" description:"Inspect modules"`
	Deep         bool     `long:"deep" description:"Enable deep check mode"`
	AwsAccessKey string   `long:"aws-access-key" description:"AWS access key used in deep check mode" value-name:"ACCESS_KEY"`
	AwsSecretKey string   `long:"aws-secret-key" description:"AWS secret key used in deep check mode" value-name:"SECRET_KEY"`
	AwsProfile   string   `long:"aws-profile" description:"AWS shared credential profile name used in deep check mode" value-name:"PROFILE"`
	AwsCredsFile string   `long:"aws-creds-file" description:"AWS shared credentials file path used in deep checking" value-name:"FILE"`
	AwsRegion    string   `long:"aws-region" description:"AWS region used in deep check mode" value-name:"REGION"`
	Force        bool     `long:"force" description:"Return zero exit status even if issues found"`
	NoColor      bool     `long:"no-color" description:"Disable colorized output"`
}

func (opts *Options) toConfig() *tflint.Config {
	ignoreModule := map[string]bool{}
	if opts.IgnoreModule != "" {
		for _, m := range strings.Split(opts.IgnoreModule, ",") {
			ignoreModule[m] = true
		}
	}

	varfile := []string{}
	if opts.Varfile != "" {
		varfile = strings.Split(opts.Varfile, ",")
	}
	if opts.Variables == nil {
		opts.Variables = []string{}
	}

	log.Printf("[DEBUG] CLI Options")
	log.Printf("[DEBUG]   Module: %t", opts.Module)
	log.Printf("[DEBUG]   DeepCheck: %t", opts.Deep)
	log.Printf("[DEBUG]   Force: %t", opts.Force)
	log.Printf("[DEBUG]   IgnoreModule: %#v", ignoreModule)
	log.Printf("[DEBUG]   EnableRules: %#v", opts.EnableRules)
	log.Printf("[DEBUG]   DisableRules: %#v", opts.DisableRules)
	log.Printf("[DEBUG]   Varfile: %#v", varfile)
	log.Printf("[DEBUG]   Variables: %#v", opts.Variables)

	rules := map[string]*tflint.RuleConfig{}
	for _, rule := range opts.EnableRules {
		rules[rule] = &tflint.RuleConfig{
			Name:    rule,
			Enabled: true,
		}
	}
	for _, rule := range opts.DisableRules {
		rules[rule] = &tflint.RuleConfig{
			Name:    rule,
			Enabled: false,
		}
	}

	return &tflint.Config{
		Module:    opts.Module,
		DeepCheck: opts.Deep,
		Force:     opts.Force,
		AwsCredentials: client.AwsCredentials{
			AccessKey: opts.AwsAccessKey,
			SecretKey: opts.AwsSecretKey,
			Profile:   opts.AwsProfile,
			CredsFile: opts.AwsCredsFile,
			Region:    opts.AwsRegion,
		},
		IgnoreModule: ignoreModule,
		Varfile:      varfile,
		Variables:    opts.Variables,
		Rules:        rules,
	}
}
