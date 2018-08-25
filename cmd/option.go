package cmd

import (
	"log"
	"strings"

	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/rules/awsrules"
	"github.com/wata727/tflint/tflint"
)

// Options is an option specified by arguments.
type Options struct {
	Version         bool   `short:"v" long:"version" description:"Print TFLint version"`
	Format          string `short:"f" long:"format" description:"Output format" choice:"default" choice:"json" choice:"checkstyle" default:"default"`
	Config          string `short:"c" long:"config" description:"Config file name" value-name:"FILE" default:".tflint.hcl"`
	IgnoreModule    string `long:"ignore-module" description:"Ignore module sources" value-name:"SOURCE1,SOURCE2..."`
	IgnoreRule      string `long:"ignore-rule" description:"Ignore rule names" value-name:"RULE1,RULE2..."`
	Varfile         string `long:"var-file" description:"Terraform variable file names" value-name:"FILE1,FILE2..."`
	Deep            bool   `long:"deep" description:"Enable deep check mode"`
	AwsAccessKey    string `long:"aws-access-key" description:"AWS access key used in deep check mode" value-name:"ACCESS_KEY"`
	AwsSecretKey    string `long:"aws-secret-key" description:"AWS secret key used in deep check mode" value-name:"SECRET_KEY"`
	AwsProfile      string `long:"aws-profile" description:"AWS shared credential profile name used in deep check mode" value-name:"PROFILE"`
	AwsRegion       string `long:"aws-region" description:"AWS region used in deep check mode" value-name:"REGION"`
	ErrorWithIssues bool   `long:"error-with-issues" description:"Return error code when issues exist"`
	Fast            bool   `long:"fast" description:"Ignore slow rules (aws_instance_invalid_ami only)"`
	Quiet           bool   `short:"q" long:"quiet" description:"Do not output any message when no issues are found (default format only)"`
}

func (opts *Options) toConfig() *tflint.Config {
	ignoreModule := map[string]bool{}
	if opts.IgnoreModule != "" {
		for _, m := range strings.Split(opts.IgnoreModule, ",") {
			ignoreModule[m] = true
		}
	}

	ignoreRule := map[string]bool{}
	if opts.IgnoreRule != "" {
		for _, r := range strings.Split(opts.IgnoreRule, ",") {
			ignoreRule[r] = true
		}
	}
	if opts.Fast {
		// `aws_instance_invalid_ami` is very slow...
		ignoreRule[awsrules.NewAwsInstanceInvalidAMIRule().Name()] = true
	}

	varfile := []string{}
	if opts.Varfile != "" {
		varfile = strings.Split(opts.Varfile, ",")
	}

	log.Printf("[DEBUG] CLI Options")
	log.Printf("[DEBUG]   DeepCheck: %t", opts.Deep)
	log.Printf("[DEBUG]   IgnoreModule: %#v", ignoreModule)
	log.Printf("[DEBUG]   IgnoreRule: %#v", ignoreRule)
	log.Printf("[DEBUG]   Varfile: %#v", varfile)

	return &tflint.Config{
		DeepCheck: opts.Deep,
		AwsCredentials: client.AwsCredentials{
			AccessKey: opts.AwsAccessKey,
			SecretKey: opts.AwsSecretKey,
			Profile:   opts.AwsProfile,
			Region:    opts.AwsRegion,
		},
		IgnoreModule:     ignoreModule,
		IgnoreRule:       ignoreRule,
		Varfile:          varfile,
		TerraformVersion: "",
		Rules:            map[string]*tflint.Rule{},
	}
}
