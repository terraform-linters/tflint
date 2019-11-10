package tflint

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/terraform-linters/tflint/client"
)

var defaultConfigFile = ".tflint.hcl"
var fallbackConfigFile = "~/.tflint.hcl"

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
	Name    string `hcl:"name,label"`
	Enabled bool   `hcl:"enabled"`
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

	return raw.toConfig(), nil
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
		ret[k] = v
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
