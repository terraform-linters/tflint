package tflint

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclparse"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/wata727/tflint/client"
)

var defaultConfigFile = ".tflint.hcl"
var fallbackConfigFile = "~/.tflint.hcl"

type rawConfig struct {
	Config *struct {
		DeepCheck        *bool              `hcl:"deep_check"`
		AwsCredentials   *map[string]string `hcl:"aws_credentials"`
		IgnoreModule     *map[string]bool   `hcl:"ignore_module"`
		IgnoreRule       *map[string]bool   `hcl:"ignore_rule"`
		Varfile          *[]string          `hcl:"varfile"`
		TerraformVersion *string            `hcl:"terraform_version"`
	} `hcl:"config,block"`
	Rules []Rule `hcl:"rule,block"`
}

// Config describes the behavior of TFLint
type Config struct {
	DeepCheck        bool
	AwsCredentials   client.AwsCredentials
	IgnoreModule     map[string]bool
	IgnoreRule       map[string]bool
	Varfile          []string
	TerraformVersion string
	Rules            map[string]*Rule
}

// Rule is a TFLint's rule config
type Rule struct {
	Name    string `hcl:"name,label"`
	Enabled bool   `hcl:"enabled"`
}

// EmptyConfig returns default config
// It is mainly used for testing
func EmptyConfig() *Config {
	return &Config{
		DeepCheck:        false,
		AwsCredentials:   client.AwsCredentials{},
		IgnoreModule:     map[string]bool{},
		IgnoreRule:       map[string]bool{},
		Varfile:          []string{},
		TerraformVersion: "",
		Rules:            map[string]*Rule{},
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

	if other.DeepCheck {
		ret.DeepCheck = true
	}

	if other.AwsCredentials.AccessKey != "" {
		ret.AwsCredentials.AccessKey = other.AwsCredentials.AccessKey
	}
	if other.AwsCredentials.SecretKey != "" {
		ret.AwsCredentials.SecretKey = other.AwsCredentials.SecretKey
	}
	if other.AwsCredentials.Profile != "" {
		ret.AwsCredentials.Profile = other.AwsCredentials.Profile
	}
	if other.AwsCredentials.Region != "" {
		ret.AwsCredentials.Region = other.AwsCredentials.Region
	}

	ret.IgnoreModule = mergeBoolMap(ret.IgnoreModule, other.IgnoreModule)
	ret.IgnoreRule = mergeBoolMap(ret.IgnoreRule, other.IgnoreRule)
	ret.Varfile = append(ret.Varfile, other.Varfile...)

	if other.TerraformVersion != "" {
		ret.TerraformVersion = other.TerraformVersion
	}

	ret.Rules = mergeRuleMap(ret.Rules, other.Rules)

	return ret
}

func (c *Config) copy() *Config {
	ignoreModule := make(map[string]bool)
	for k, v := range c.IgnoreModule {
		ignoreModule[k] = v
	}

	ignoreRule := make(map[string]bool)
	for k, v := range c.IgnoreRule {
		ignoreRule[k] = v
	}

	varfile := make([]string, len(c.Varfile))
	copy(varfile, c.Varfile)

	rules := map[string]*Rule{}
	for k, v := range c.Rules {
		rules[k] = &Rule{}
		*rules[k] = *v
	}

	return &Config{
		DeepCheck:        c.DeepCheck,
		AwsCredentials:   c.AwsCredentials,
		IgnoreModule:     ignoreModule,
		IgnoreRule:       ignoreRule,
		Varfile:          varfile,
		TerraformVersion: c.TerraformVersion,
		Rules:            rules,
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

	cfg := raw.toConfig()
	log.Printf("[DEBUG] Config loaded")
	log.Printf("[DEBUG]   DeepCheck: %t", cfg.DeepCheck)
	log.Printf("[DEBUG]   IgnoreModule: %#v", cfg.IgnoreModule)
	log.Printf("[DEBUG]   IgnoreRule: %#v", cfg.IgnoreRule)
	log.Printf("[DEBUG]   Varfile: %#v", cfg.Varfile)
	log.Printf("[DEBUG]   TerraformVersion: %s", cfg.TerraformVersion)
	log.Printf("[DEBUG]   Rules: %#v", cfg.Rules)

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

func mergeRuleMap(a, b map[string]*Rule) map[string]*Rule {
	ret := map[string]*Rule{}
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
		if rc.DeepCheck != nil {
			ret.DeepCheck = *rc.DeepCheck
		}
		if rc.AwsCredentials != nil {
			credentials := *rc.AwsCredentials
			ret.AwsCredentials.AccessKey = credentials["access_key"]
			ret.AwsCredentials.SecretKey = credentials["secret_key"]
			ret.AwsCredentials.Profile = credentials["profile"]
			ret.AwsCredentials.Region = credentials["region"]
		}
		if rc.IgnoreModule != nil {
			ret.IgnoreModule = *rc.IgnoreModule
		}
		if rc.IgnoreRule != nil {
			ret.IgnoreRule = *rc.IgnoreRule
		}
		if rc.Varfile != nil {
			ret.Varfile = *rc.Varfile
		}
		if rc.TerraformVersion != nil {
			ret.TerraformVersion = *rc.TerraformVersion
		}
	}

	for _, r := range raw.Rules {
		var rule = r
		ret.Rules[rule.Name] = &rule
	}

	return ret
}
