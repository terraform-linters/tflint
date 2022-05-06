package tflint

import (
	"fmt"
	"log"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

var defaultConfigFile = ".tflint.hcl"
var fallbackConfigFile = "~/.tflint.hcl"

var configSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "config",
		},
		{
			Type:       "rule",
			LabelNames: []string{"name"},
		},
		{
			Type:       "plugin",
			LabelNames: []string{"name"},
		},
	},
}

var innerConfigSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "module"},
		{Name: "force"},
		{Name: "ignore_module"},
		{Name: "varfile"},
		{Name: "variables"},
		{Name: "disabled_by_default"},
		{Name: "plugin_dir"},
		{Name: "format"},
	},
}

var validFormats = []string{
	"default",
	"json",
	"checkstyle",
	"junit",
	"compact",
	"sarif",
}

// Config describes the behavior of TFLint
type Config struct {
	Module            bool
	Force             bool
	IgnoreModules     map[string]bool
	Varfiles          []string
	Variables         []string
	DisabledByDefault bool
	PluginDir         string
	Format            string
	Rules             map[string]*RuleConfig
	Plugins           map[string]*PluginConfig

	sources map[string][]byte
}

// RuleConfig is a TFLint's rule config
type RuleConfig struct {
	Name    string   `hcl:"name,label"`
	Enabled bool     `hcl:"enabled"`
	Body    hcl.Body `hcl:",remain"`
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

// LoadConfig reads TFLint config from a file.
// If ./.tflint.hcl does not exist, load ~/.tflint.hcl.
// This fallback does not fire when explicitly reading a file other than .tflint.hcl.
func LoadConfig(fs afero.Afero, file string) (*Config, error) {
	log.Printf("[INFO] Load config: %s", file)
	if f, err := fs.Open(file); err == nil {
		cfg, err := loadConfig(f)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if file != defaultConfigFile {
		return nil, fmt.Errorf("failed to load file: %w", err)
	} else {
		log.Printf("[INFO] file not found")
	}

	fallback, err := homedir.Expand(fallbackConfigFile)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Load config: %s", fallback)
	if f, err := fs.Open(fallback); err == nil {
		cfg, err := loadConfig(f)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	log.Printf("[INFO] file not found")

	log.Print("[INFO] Use default config")
	return EmptyConfig(), nil
}

func loadConfig(file afero.File) (*Config, error) {
	src, err := afero.ReadAll(file)
	if err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(src, file.Name())
	if diags.HasErrors() {
		return nil, diags
	}

	content, diags := f.Body.Content(configSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	config := EmptyConfig()
	config.sources = parser.Sources()
	for _, block := range content.Blocks {
		switch block.Type {
		case "config":
			inner, diags := block.Body.Content(innerConfigSchema)
			if diags.HasErrors() {
				return config, diags
			}

			for name, attr := range inner.Attributes {
				switch name {
				case "module":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.Module); err != nil {
						return config, err
					}
				case "force":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.Force); err != nil {
						return config, err
					}
				case "ignore_module":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.IgnoreModules); err != nil {
						return config, err
					}
				case "varfile":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.Varfiles); err != nil {
						return config, err
					}
				case "variables":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.Variables); err != nil {
						return config, err
					}
				case "disabled_by_default":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.DisabledByDefault); err != nil {
						return config, err
					}
				case "plugin_dir":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.PluginDir); err != nil {
						return config, err
					}
				case "format":
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.Format); err != nil {
						return config, err
					}
					formatValid := false
					for _, f := range validFormats {
						if config.Format == "" || config.Format == f {
							formatValid = true
							break
						}
					}
					if !formatValid {
						return config, fmt.Errorf("%s is invalid format. Allowed formats are: %s", config.Format, strings.Join(validFormats, ", "))
					}
				default:
					panic("never happened")
				}
			}
		case "rule":
			ruleConfig := &RuleConfig{Name: block.Labels[0]}
			if err := gohcl.DecodeBody(block.Body, nil, ruleConfig); err != nil {
				return config, err
			}
			config.Rules[block.Labels[0]] = ruleConfig
		case "plugin":
			pluginConfig := &PluginConfig{Name: block.Labels[0]}
			if err := gohcl.DecodeBody(block.Body, nil, pluginConfig); err != nil {
				return config, err
			}
			if err := pluginConfig.validate(); err != nil {
				return config, err
			}
			config.Plugins[block.Labels[0]] = pluginConfig
		default:
			panic("never happened")
		}
	}

	log.Printf("[DEBUG] Config loaded")
	log.Printf("[DEBUG]   Module: %t", config.Module)
	log.Printf("[DEBUG]   Force: %t", config.Force)
	log.Printf("[DEBUG]   IgnoreModules:")
	for name, ignore := range config.IgnoreModules {
		log.Printf("[DEBUG]     %s: %t", name, ignore)
	}
	log.Printf("[DEBUG]   Varfiles: %s", strings.Join(config.Varfiles, ", "))
	log.Printf("[DEBUG]   Variables: %s", strings.Join(config.Variables, ", "))
	log.Printf("[DEBUG]   DisabledByDefault: %t", config.DisabledByDefault)
	log.Printf("[DEBUG]   PluginDir: %s", config.PluginDir)
	log.Printf("[DEBUG]   Format: %s", config.Format)
	log.Printf("[DEBUG]   Rules:")
	for name, rule := range config.Rules {
		log.Printf("[DEBUG]     %s: %t", name, rule.Enabled)
	}
	log.Printf("[DEBUG]   Plugins:")
	for name, plugin := range config.Plugins {
		log.Printf("[DEBUG]     %s: enabled=%t, version=%s, source=%s", name, plugin.Enabled, plugin.Version, plugin.Source)
	}

	return config, nil
}

// Sources returns parsed config file sources.
// Normally, there is only one file, but it is represented by map to retain the file name.
func (c *Config) Sources() map[string][]byte {
	return c.sources
}

// Merge merges the two configs and applies to itself.
// Since the argument takes precedence, it can be used as overwriting of the config.
func (c *Config) Merge(other *Config) {
	if other.Module {
		c.Module = true
	}
	if other.Force {
		c.Force = true
	}
	if other.DisabledByDefault {
		c.DisabledByDefault = true
	}
	if other.PluginDir != "" {
		c.PluginDir = other.PluginDir
	}
	if other.Format != "" {
		c.Format = other.Format
	}

	for name, ignore := range other.IgnoreModules {
		c.IgnoreModules[name] = ignore
	}
	c.Varfiles = append(c.Varfiles, other.Varfiles...)
	c.Variables = append(c.Variables, other.Variables...)

	for name, rule := range other.Rules {
		// HACK: If you enable the rule through the CLI instead of the file, its hcl.Body will be nil.
		//       In this case, only override Enabled flag
		if _, exists := c.Rules[name]; exists && rule.Body == nil {
			c.Rules[name].Enabled = rule.Enabled
		} else {
			c.Rules[name] = rule
		}
	}

	for name, plugin := range other.Plugins {
		// HACK: If you enable the plugin through the CLI instead of the file, its hcl.Body will be nil.
		//       In this case, only override Enabled flag
		if _, exists := c.Plugins[name]; exists && plugin.Body == nil {
			c.Plugins[name].Enabled = plugin.Enabled
		} else {
			c.Plugins[name] = plugin
		}
	}
}

// ToPluginConfig converts self into the plugin configuration format
func (c *Config) ToPluginConfig() *sdk.Config {
	cfg := &sdk.Config{
		Rules:             map[string]*sdk.RuleConfig{},
		DisabledByDefault: c.DisabledByDefault,
	}
	for _, rule := range c.Rules {
		cfg.Rules[rule.Name] = &sdk.RuleConfig{
			Name:    rule.Name,
			Enabled: rule.Enabled,
		}
	}
	return cfg
}

// Content extracts a plugin config based on the passed schema.
func (c *PluginConfig) Content(schema *hclext.BodySchema) (*hclext.BodyContent, hcl.Diagnostics) {
	if schema == nil {
		schema = &hclext.BodySchema{}
	}
	if c.Body == nil {
		return &hclext.BodyContent{}, hcl.Diagnostics{}
	}
	return hclext.Content(c.Body, schema)
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
		rulesetName, err := ruleset.RuleSetName()
		if err != nil {
			return err
		}
		ruleNames, err := ruleset.RuleNames()
		if err != nil {
			return err
		}

		for _, rule := range ruleNames {
			if existsName, exists := rulesMap[rule]; exists {
				return fmt.Errorf("`%s` is duplicated in %s and %s", rule, existsName, rulesetName)
			}
			rulesMap[rule] = rulesetName
		}
	}

	for _, rule := range c.Rules {
		if _, exists := rulesMap[rule.Name]; !exists {
			return fmt.Errorf("Rule not found: %s", rule.Name)
		}
	}

	return nil
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
