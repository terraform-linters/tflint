package tflint

import (
	"fmt"
	"log"
	"os"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	Module    bool
	ModuleSet bool

	Force    bool
	ForceSet bool

	DisabledByDefault    bool
	DisabledByDefaultSet bool

	PluginDir    string
	PluginDirSet bool

	Format    string
	FormatSet bool

	Varfiles      []string
	Variables     []string
	Only          []string
	IgnoreModules map[string]bool
	Rules         map[string]*RuleConfig
	Plugins       map[string]*PluginConfig

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
	SourceHost  string
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

// LoadConfig loads TFLint config file.
// The priority of the configuration files is as follows:
//
// 1. file passed by the --config option
// 2. file set by the TFLINT_CONFIG_FILE environment variable
// 3. current directory (./.tflint.hcl)
// 4. home directory (~/.tflint.hcl)
//
// For 1 and 2, if the file does not exist, an error will be returned immediately.
// If 3 fails, fallback to 4, and If it fails, an empty configuration is returned.
//
// It also automatically enables bundled plugin if the "terraform"
// plugin block is not explicitly declared.
func LoadConfig(fs afero.Afero, file string) (*Config, error) {
	// Load the file passed by the --config option
	if file != defaultConfigFile {
		log.Printf("[INFO] Load config: %s", file)
		f, err := fs.Open(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load file: %w", err)
		}
		cfg, err := loadConfig(f)
		if err != nil {
			return nil, err
		}
		return cfg.enableBundledPlugin(), nil
	}

	// Load the file set by the environment variable
	envFile := os.Getenv("TFLINT_CONFIG_FILE")
	if envFile != "" {
		log.Printf("[INFO] Found TFLINT_CONFIG_FILE. Load config: %s", envFile)
		f, err := fs.Open(envFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load file: %w", err)
		}
		cfg, err := loadConfig(f)
		if err != nil {
			return nil, err
		}
		return cfg.enableBundledPlugin(), nil
	}

	// Load the default config file
	log.Printf("[INFO] Load config: %s", file)
	if f, err := fs.Open(file); err == nil {
		cfg, err := loadConfig(f)
		if err != nil {
			return nil, err
		}
		return cfg.enableBundledPlugin(), nil
	}
	log.Printf("[INFO] file not found")

	// Load the fallback config file
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
		return cfg.enableBundledPlugin(), nil
	}
	log.Printf("[INFO] file not found")

	// Use the default config
	log.Print("[INFO] Use default config")
	return EmptyConfig().enableBundledPlugin(), nil
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
					config.ModuleSet = true
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.Module); err != nil {
						return config, err
					}
				case "force":
					config.ForceSet = true
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
					config.DisabledByDefaultSet = true
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.DisabledByDefault); err != nil {
						return config, err
					}
				case "plugin_dir":
					config.PluginDirSet = true
					if err := gohcl.DecodeExpression(attr.Expr, nil, &config.PluginDir); err != nil {
						return config, err
					}
				case "format":
					config.FormatSet = true
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
	log.Printf("[DEBUG]   ModuleSet: %t", config.ModuleSet)
	log.Printf("[DEBUG]   Force: %t", config.Force)
	log.Printf("[DEBUG]   ForceSet: %t", config.ForceSet)
	log.Printf("[DEBUG]   DisabledByDefault: %t", config.DisabledByDefault)
	log.Printf("[DEBUG]   DisabledByDefaultSet: %t", config.DisabledByDefaultSet)
	log.Printf("[DEBUG]   PluginDir: %s", config.PluginDir)
	log.Printf("[DEBUG]   PluginDirSet: %t", config.PluginDirSet)
	log.Printf("[DEBUG]   Format: %s", config.Format)
	log.Printf("[DEBUG]   FormatSet: %t", config.FormatSet)
	log.Printf("[DEBUG]   Varfiles: %s", strings.Join(config.Varfiles, ", "))
	log.Printf("[DEBUG]   Variables: %s", strings.Join(config.Variables, ", "))
	log.Printf("[DEBUG]   Only: %s", strings.Join(config.Only, ", "))
	log.Printf("[DEBUG]   IgnoreModules:")
	for name, ignore := range config.IgnoreModules {
		log.Printf("[DEBUG]     %s: %t", name, ignore)
	}
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

// Enable the "recommended" preset if the bundled plugin is automatically enabled.
var bundledPluginConfigFilename = "__bundled_plugin_config.hcl"
var bundledPluginConfigContent = `
preset = "recommended"
`

// DisbaleBundledPlugin is a flag to temporarily disable the bundled plugin for integration tests.
var DisableBundledPlugin = false

// Terraform Language plugin is automatically enabled if the plugin isn't explicitly declared.
func (c *Config) enableBundledPlugin() *Config {
	if DisableBundledPlugin {
		return c
	}

	f, diags := hclsyntax.ParseConfig([]byte(bundledPluginConfigContent), bundledPluginConfigFilename, hcl.InitialPos)
	if diags.HasErrors() {
		panic(diags)
	}

	if _, exists := c.Plugins["terraform"]; !exists {
		log.Print("[INFO] The `terraform` plugin block is not found. Enable the plugin `terraform` automatically")

		c.Plugins["terraform"] = &PluginConfig{
			Name:    "terraform",
			Enabled: true,
			Body:    f.Body,
		}

		// Implicit preset is ignored if you enable DisabledByDefault
		if c.DisabledByDefault {
			c.Plugins["terraform"].Body = nil
		}
	}
	return c
}

// Sources returns parsed config file sources.
// To support bundle plugin config, this function returns c.sources
// with a merge of the pseudo config file.
func (c *Config) Sources() map[string][]byte {
	ret := map[string][]byte{
		bundledPluginConfigFilename: []byte(bundledPluginConfigContent),
	}

	for name, content := range c.sources {
		ret[name] = content
	}
	return ret
}

// Merge merges the two configs and applies to itself.
// Since the argument takes precedence, it can be used as overwriting of the config.
func (c *Config) Merge(other *Config) {
	if other.ModuleSet {
		c.ModuleSet = true
		c.Module = other.Module
	}
	if other.ForceSet {
		c.ForceSet = true
		c.Force = other.Force
	}
	if other.DisabledByDefaultSet {
		c.DisabledByDefaultSet = true
		c.DisabledByDefault = other.DisabledByDefault
	}
	if other.PluginDirSet {
		c.PluginDirSet = true
		c.PluginDir = other.PluginDir
	}
	if other.FormatSet {
		c.FormatSet = true
		c.Format = other.Format
	}

	c.Varfiles = append(c.Varfiles, other.Varfiles...)
	c.Variables = append(c.Variables, other.Variables...)
	c.Only = append(c.Only, other.Only...)

	for name, ignore := range other.IgnoreModules {
		c.IgnoreModules[name] = ignore
	}

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
		Only:              c.Only,
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

// RuleSet is an interface to handle plugin's RuleSet.
// The real impl is github.com/terraform-linters/tflint-plugin-sdk/plugin/host2plugin.GRPCClient.
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
			return fmt.Errorf("plugin `%s`: `source` is invalid. Must be a GitHub reference in the format `${host}/${owner}/${repo}`", c.Name)
		}

		c.SourceHost = parts[0]
		c.SourceOwner = parts[1]
		c.SourceRepo = parts[2]
	}

	return nil
}
