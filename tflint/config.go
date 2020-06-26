package tflint

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	homedir "github.com/mitchellh/go-homedir"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/client"
)

var defaultConfigFile = ".tflint.hcl"
var fallbackConfigFile = "~/.tflint.hcl"

var removedRulesMap = map[string]string{
	"terraform_dash_in_data_source_name": "`terraform_dash_in_data_source_name` rule was removed in v0.16.0. Please use `terraform_naming_convention` rule instead",
	"terraform_dash_in_module_name":      "`terraform_dash_in_module_name` rule was removed in v0.16.0. Please use `terraform_naming_convention` rule instead",
	"terraform_dash_in_output_name":      "`terraform_dash_in_output_name` rule was removed in v0.16.0. Please use `terraform_naming_convention` rule instead",
	"terraform_dash_in_resource_name":    "`terraform_dash_in_resource_name` rule was removed in v0.16.0. Please use `terraform_naming_convention` rule instead",
}

type rawConfig struct {
	Config *struct {
		Module         *bool              `hcl:"module"`
		DeepCheck      *bool              `hcl:"deep_check"`
		Force          *bool              `hcl:"force"`
		AwsCredentials *map[string]string `hcl:"aws_credentials"`
		IgnoreModule   *map[string]bool   `hcl:"ignore_module"`
		Varfile        *[]string          `hcl:"varfile"`
		Variables      *[]string          `hcl:"variables"`
		// Removed options
		TerraformVersion *string          `hcl:"terraform_version"`
		IgnoreRule       *map[string]bool `hcl:"ignore_rule"`
	} `hcl:"config,block"`
	Rules   []RuleConfig   `hcl:"rule,block"`
	Plugins []PluginConfig `hcl:"plugin,block"`
}

// Config describes the behavior of TFLint
type Config struct {
	Module         bool
	DeepCheck      bool
	Force          bool
	AwsCredentials client.AwsCredentials
	IgnoreModules  map[string]bool
	Varfiles       []string
	Variables      []string
	Rules          map[string]*RuleConfig
	Plugins        map[string]*PluginConfig
}

// RuleConfig is a TFLint's rule config
type RuleConfig struct {
	Name    string   `hcl:"name,label"`
	Enabled bool     `hcl:"enabled"`
	Body    hcl.Body `hcl:",remain"`
}

// PluginConfig is a TFLint's plugin config
type PluginConfig struct {
	Name    string `hcl:"name,label"`
	Enabled bool   `hcl:"enabled"`
}

// EmptyConfig returns default config
// It is mainly used for testing
func EmptyConfig() *Config {
	return &Config{
		Module:         false,
		DeepCheck:      false,
		Force:          false,
		AwsCredentials: client.AwsCredentials{},
		IgnoreModules:  map[string]bool{},
		Varfiles:       []string{},
		Variables:      []string{},
		Rules:          map[string]*RuleConfig{},
		Plugins:        map[string]*PluginConfig{},
	}
}

// LoadConfig loads TFLint config from file
// If failed to load the default config file, it tries to load config file under the home directory
// Therefore, if there is no default config file, it will not return an error
func LoadConfig(dir string) (*Config, error) {
	log.Printf("[INFO] Load and merge config: %s", dir)
	config, err := loadMergedConfig(dir)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}

	if config != nil {
		return config, nil
	}

	fallback, err := homedir.Expand(fallbackConfigFile)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}

	log.Printf("[INFO] Load fallback config: %s", fallback)
	if _, err := os.Stat(fallback); !os.IsNotExist(err) {
		cfg, err := loadConfigFromFile(fallback)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	log.Printf("[INFO] Fallback config file is not found. Ignored")

	log.Print("[INFO] Use default config")
	return EmptyConfig(), nil
}

func loadMergedConfig(dir string) (*Config, error) {
	configs := make([]*Config, 0)
	for ; dir != "/"; dir = filepath.Dir(dir) {
		file := filepath.Join(dir, defaultConfigFile)

		log.Printf("[DEBUG] Checking for config: %s", file)
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			cfg, err := loadConfigFromFile(file)
			if err != nil {
				return nil, err
			}

			log.Printf("[DEBUG] Found config: %s", file)
			configs = append(configs, cfg)
		}
	}

	// reverse
	for i := len(configs)/2 - 1; i >= 0; i-- {
		opp := len(configs) - 1 - i
		configs[i], configs[opp] = configs[opp], configs[i]
	}

	if len(configs) == 0 {
		return nil, nil
	}

	cfg := EmptyConfig()
	for _, c := range configs {
		cfg = cfg.Merge(c)
	}
	return cfg, nil
}

// Merge returns a merged copy of the two configs
// Since the argument takes precedence, it can be used as overwriting of the config
func (c *Config) Merge(other *Config) *Config {
	ret := c.copy()

	if other.Module {
		ret.Module = true
	}
	if other.DeepCheck {
		ret.DeepCheck = true
	}
	if other.Force {
		ret.Force = true
	}

	ret.AwsCredentials = ret.AwsCredentials.Merge(other.AwsCredentials)
	ret.IgnoreModules = mergeBoolMap(ret.IgnoreModules, other.IgnoreModules)
	ret.Varfiles = append(ret.Varfiles, other.Varfiles...)
	ret.Variables = append(ret.Variables, other.Variables...)

	ret.Rules = mergeRuleMap(ret.Rules, other.Rules)
	ret.Plugins = mergePluginMap(ret.Plugins, other.Plugins)

	return ret
}

// ToPluginConfig converts self into the plugin configuration format
func (c *Config) ToPluginConfig() *tfplugin.Config {
	cfg := &tfplugin.Config{Rules: map[string]*tfplugin.RuleConfig{}}
	for _, rule := range c.Rules {
		cfg.Rules[rule.Name] = &tfplugin.RuleConfig{
			Name:    rule.Name,
			Enabled: rule.Enabled,
		}
	}
	return cfg
}

// RuleSet is an interface to handle plugin's RuleSet and core RuleSet both
// In the future, when all RuleSets are cut out into plugins, it will no longer be needed.
type RuleSet interface {
	RuleSetName() (string, error)
	RuleSetVersion() (string, error)
	RuleNames() ([]string, error)
}

// ValidateRules checks for duplicate rule names, for invalid rule names, and so on.
func (c *Config) ValidateRules(rulesets ...RuleSet) error {
	rulesMap := map[string]string{}
	for _, ruleset := range rulesets {
		ruleNames, err := ruleset.RuleNames()
		if err != nil {
			return err
		}

		for _, rule := range ruleNames {
			rulesetName, err := ruleset.RuleSetName()
			if err != nil {
				return err
			}

			if existsName, exists := rulesMap[rule]; exists {
				return fmt.Errorf("`%s` is duplicated in %s and %s", rule, existsName, rulesetName)
			}
			rulesMap[rule] = rulesetName
		}
	}

	for _, rule := range c.Rules {
		if _, exists := rulesMap[rule.Name]; !exists {
			if message, exists := removedRulesMap[rule.Name]; exists {
				return errors.New(message)
			}
			return fmt.Errorf("Rule not found: %s", rule.Name)
		}
	}

	return nil
}

func (c *Config) copy() *Config {
	ignoreModules := make(map[string]bool)
	for k, v := range c.IgnoreModules {
		ignoreModules[k] = v
	}

	varfiles := make([]string, len(c.Varfiles))
	copy(varfiles, c.Varfiles)

	variables := make([]string, len(c.Variables))
	copy(variables, c.Variables)

	rules := map[string]*RuleConfig{}
	for k, v := range c.Rules {
		rules[k] = &RuleConfig{}
		*rules[k] = *v
	}

	plugins := map[string]*PluginConfig{}
	for k, v := range c.Plugins {
		plugins[k] = &PluginConfig{}
		*plugins[k] = *v
	}

	return &Config{
		Module:         c.Module,
		DeepCheck:      c.DeepCheck,
		Force:          c.Force,
		AwsCredentials: c.AwsCredentials,
		IgnoreModules:  ignoreModules,
		Varfiles:       varfiles,
		Variables:      variables,
		Rules:          rules,
		Plugins:        plugins,
	}
}

func loadConfigFromFile(file string) (*Config, error) {
	parser := hclparse.NewParser()

	f, diags := parser.ParseHCLFile(file)
	if diags.HasErrors() {
		return nil, diags
	}

	var raw rawConfig
	diags = gohcl.DecodeBody(f.Body, nil, &raw)
	if diags.HasErrors() {
		return nil, diags
	}

	if raw.Config != nil {
		if raw.Config.TerraformVersion != nil {
			return nil, errors.New("`terraform_version` was removed in v0.9.0 because the option is no longer used")
		}

		if raw.Config.IgnoreRule != nil {
			return nil, errors.New("`ignore_rule` was removed in v0.12.0. Please define `rule` block with `enabled = false` instead")
		}
	}

	cfg := raw.toConfig()
	log.Printf("[DEBUG] Config loaded")
	log.Printf("[DEBUG]   Module: %t", cfg.Module)
	log.Printf("[DEBUG]   DeepCheck: %t", cfg.DeepCheck)
	log.Printf("[DEBUG]   Force: %t", cfg.Force)
	log.Printf("[DEBUG]   IgnoreModules: %#v", cfg.IgnoreModules)
	log.Printf("[DEBUG]   Varfiles: %#v", cfg.Varfiles)
	log.Printf("[DEBUG]   Variables: %#v", cfg.Variables)
	log.Printf("[DEBUG]   Rules: %#v", cfg.Rules)
	log.Printf("[DEBUG]   Plugins: %#v", cfg.Plugins)

	return cfg, nil
}

func mergeBoolMap(a, b map[string]bool) map[string]bool {
	ret := map[string]bool{}
	for k, v := range a {
		ret[k] = v
	}
	for k, v := range b {
		ret[k] = v
	}
	return ret
}

func mergeRuleMap(a, b map[string]*RuleConfig) map[string]*RuleConfig {
	ret := map[string]*RuleConfig{}
	for k, v := range a {
		ret[k] = v
	}
	for k, v := range b {
		// HACK: If you enable the rule through the CLI instead of the file, its hcl.Body will not contain valid range.
		// @see https://github.com/hashicorp/hcl/blob/v2.5.0/merged.go#L132-L135
		if prevConfig, exists := ret[k]; exists && v.Body.MissingItemRange().Filename == "<empty>" {
			ret[k] = v
			// Do not override body
			ret[k].Body = prevConfig.Body
		} else {
			ret[k] = v
		}
	}
	return ret
}

func mergePluginMap(a, b map[string]*PluginConfig) map[string]*PluginConfig {
	ret := map[string]*PluginConfig{}
	for k, v := range a {
		ret[k] = v
	}
	for k, v := range b {
		ret[k] = v
	}
	return ret
}

func (raw *rawConfig) toConfig() *Config {
	ret := EmptyConfig()
	rc := raw.Config

	if rc != nil {
		if rc.Module != nil {
			ret.Module = *rc.Module
		}
		if rc.DeepCheck != nil {
			ret.DeepCheck = *rc.DeepCheck
		}
		if rc.Force != nil {
			ret.Force = *rc.Force
		}
		if rc.AwsCredentials != nil {
			credentials := *rc.AwsCredentials
			ret.AwsCredentials.AccessKey = credentials["access_key"]
			ret.AwsCredentials.SecretKey = credentials["secret_key"]
			ret.AwsCredentials.Profile = credentials["profile"]
			ret.AwsCredentials.CredsFile = credentials["shared_credentials_file"]
			ret.AwsCredentials.Region = credentials["region"]
		}
		if rc.IgnoreModule != nil {
			ret.IgnoreModules = *rc.IgnoreModule
		}
		if rc.Varfile != nil {
			ret.Varfiles = *rc.Varfile
		}
		if rc.Variables != nil {
			ret.Variables = *rc.Variables
		}
	}

	for _, r := range raw.Rules {
		var rule = r
		ret.Rules[rule.Name] = &rule
	}

	for _, p := range raw.Plugins {
		var plugin = p
		ret.Plugins[plugin.Name] = &plugin
	}

	return ret
}
