package tflint

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	homedir "github.com/mitchellh/go-homedir"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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
		Module            *bool            `hcl:"module"`
		Force             *bool            `hcl:"force"`
		IgnoreModule      *map[string]bool `hcl:"ignore_module"`
		Varfile           *[]string        `hcl:"varfile"`
		Variables         *[]string        `hcl:"variables"`
		DisabledByDefault *bool            `hcl:"disabled_by_default"`
		// Removed options
		TerraformVersion *string            `hcl:"terraform_version"`
		IgnoreRule       *map[string]bool   `hcl:"ignore_rule"`
		DeepCheck        *bool              `hcl:"deep_check"`
		AwsCredentials   *map[string]string `hcl:"aws_credentials"`
	} `hcl:"config,block"`
	Rules   []RuleConfig   `hcl:"rule,block"`
	Plugins []PluginConfig `hcl:"plugin,block"`
}

// Config describes the behavior of TFLint
type Config struct {
	Module            bool
	Force             bool
	IgnoreModules     map[string]bool
	Varfiles          []string
	Variables         []string
	DisabledByDefault bool
	Rules             map[string]*RuleConfig
	Plugins           map[string]*PluginConfig
}

// RuleConfig is a TFLint's rule config
type RuleConfig struct {
	Name    string   `hcl:"name,label"`
	Enabled bool     `hcl:"enabled"`
	Body    hcl.Body `hcl:",remain"`

	// file is the result of parsing the HCL file that declares the rule configuration.
	file *hcl.File
}

// PluginConfig is a TFLint's plugin config
type PluginConfig struct {
	Name       string `hcl:"name,label"`
	Enabled    bool   `hcl:"enabled"`
	Version    string `hcl:"version,optional"`
	Source     string `hcl:"source,optional"`
	SigningKey string `hcl:"signing_key,optional"`

	Body hcl.Body `hcl:",remain"`

	// Parsed source attributes
	SourceOwner string
	SourceRepo  string

	// file is the result of parsing the HCL file that declares the plugin configuration.
	file *hcl.File
}

// EmptyConfig returns default config
// It is mainly used for testing
func EmptyConfig() *Config {
	return &Config{
		Module:            false,
		Force:             false,
		IgnoreModules:     map[string]bool{},
		Varfiles:          []string{},
		Variables:         []string{},
		DisabledByDefault: false,
		Rules:             map[string]*RuleConfig{},
		Plugins:           map[string]*PluginConfig{},
	}
}

// LoadConfig loads TFLint config from file
// If failed to load the default config file, it tries to load config file under the home directory
// Therefore, if there is no default config file, it will not return an error
func LoadConfig(file string) (*Config, error) {
	log.Printf("[INFO] Load config: %s", file)
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		cfg, err := loadConfigFromFile(file)
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return nil, err
		}
		return cfg, nil
	} else if file != defaultConfigFile {
		log.Printf("[ERROR] %s", err)
		return nil, fmt.Errorf("`%s` is not found", file)
	} else {
		log.Printf("[INFO] Default config file is not found. Ignored")
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

// Merge returns a merged copy of the two configs
// Since the argument takes precedence, it can be used as overwriting of the config
func (c *Config) Merge(other *Config) *Config {
	ret := c.copy()

	if other.Module {
		ret.Module = true
	}
	if other.Force {
		ret.Force = true
	}
	if other.DisabledByDefault {
		ret.DisabledByDefault = true
	}

	ret.IgnoreModules = mergeBoolMap(ret.IgnoreModules, other.IgnoreModules)
	ret.Varfiles = append(ret.Varfiles, other.Varfiles...)
	ret.Variables = append(ret.Variables, other.Variables...)

	ret.Rules = mergeRuleMap(ret.Rules, other.Rules)
	ret.Plugins = mergePluginMap(ret.Plugins, other.Plugins)

	return ret
}

// ToPluginConfig converts self into the plugin configuration format
func (c *Config) ToPluginConfig(name string) *tfplugin.MarshalledConfig {
	pluginCfg := c.Plugins[name]

	var bodyBytes []byte
	var cfgRange hcl.Range
	if pluginCfg.Body != nil {
		cfgRange = configBodyRange(pluginCfg.Body)
		bodyBytes = cfgRange.SliceBytes(pluginCfg.file.Bytes)
	}

	cfg := &tfplugin.MarshalledConfig{
		Rules:             map[string]*tfplugin.RuleConfig{},
		DisabledByDefault: c.DisabledByDefault,
		BodyBytes:         bodyBytes,
		BodyRange:         cfgRange,
	}
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
		Module:            c.Module,
		Force:             c.Force,
		IgnoreModules:     ignoreModules,
		Varfiles:          varfiles,
		Variables:         variables,
		DisabledByDefault: c.DisabledByDefault,
		Rules:             rules,
		Plugins:           plugins,
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

		if raw.Config.DeepCheck != nil {
			return nil, errors.New(`global "deep_check" option was removed in v0.23.0. Please declare "plugin" block like the following:

plugin "aws" {
  enabled = true
  deep_check = true
}`)
		}

		if raw.Config.AwsCredentials != nil {
			return nil, errors.New(`"aws_credentials" was removed in v0.23.0. Please declare "plugin" block like the following:

plugin "aws" {
  enabled = true
  deep_check = true
  access_key = ...
}`)
		}
	}

	cfg := raw.toConfig()
	for _, rule := range cfg.Rules {
		rule.file = f
	}
	for _, plugin := range cfg.Plugins {
		plugin.file = f

		if err := plugin.validate(); err != nil {
			return nil, err
		}
	}

	log.Printf("[DEBUG] Config loaded")
	log.Printf("[DEBUG]   Module: %t", cfg.Module)
	log.Printf("[DEBUG]   Force: %t", cfg.Force)
	log.Printf("[DEBUG]   IgnoreModules: %#v", cfg.IgnoreModules)
	log.Printf("[DEBUG]   Varfiles: %#v", cfg.Varfiles)
	log.Printf("[DEBUG]   Variables: %#v", cfg.Variables)
	log.Printf("[DEBUG]   DisabledByDefault: %#v", cfg.DisabledByDefault)
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
			// Do not override body and file
			ret[k].Body = prevConfig.Body
			ret[k].file = prevConfig.file
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
		if prevConfig, exists := ret[k]; exists && v.Body == nil {
			// Use the plugin config from the config file when the plugin is enabled via CLI
			ret[k].Body = prevConfig.Body
			ret[k].file = prevConfig.file
		}
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
		if rc.Force != nil {
			ret.Force = *rc.Force
		}
		if rc.DisabledByDefault != nil {
			ret.DisabledByDefault = *rc.DisabledByDefault
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

// Bytes returns the bytes of the configuration file declared in the rule.
func (r *RuleConfig) Bytes() []byte {
	return r.file.Bytes
}

func configBodyRange(body hcl.Body) hcl.Range {
	var bodyRange hcl.Range

	// Estimate the range of the body from the range of all attributes and blocks.
	hclBody := body.(*hclsyntax.Body)
	for _, attr := range hclBody.Attributes {
		if bodyRange.Empty() {
			bodyRange = attr.Range()
		} else {
			bodyRange = hcl.RangeOver(bodyRange, attr.Range())
		}
	}
	for _, block := range hclBody.Blocks {
		if bodyRange.Empty() {
			bodyRange = block.Range()
		} else {
			bodyRange = hcl.RangeOver(bodyRange, block.Range())
		}
	}
	return bodyRange
}

func (c *PluginConfig) validate() error {
	if c.Version != "" && c.Source == "" {
		return fmt.Errorf("plugin `%s`: `source` attribute cannot be omitted when specifying `version`", c.Name)
	}

	if c.Source != "" {
		if c.Version == "" {
			return fmt.Errorf("plugin `%s`: `version` attribute cannot be omitted when specifying `source`", c.Name)
		}

		parts := strings.Split(c.Source, "/")
		// Expected `github.com/owner/repo` format
		if len(parts) != 3 {
			return fmt.Errorf("plugin `%s`: `source` is invalid. Must be in the format `github.com/owner/repo`", c.Name)
		}
		if parts[0] != "github.com" {
			return fmt.Errorf("plugin `%s`: `source` is invalid. Hostname must be `github.com`", c.Name)
		}
		c.SourceOwner = parts[1]
		c.SourceRepo = parts[2]
	}

	return nil
}
