package config

import (
	"os"
	"strings"

	"io/ioutil"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclparse"
)

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

type Config struct {
	Debug              bool
	DeepCheck          bool
	AwsCredentials     map[string]string
	IgnoreModule       map[string]bool
	IgnoreRule         map[string]bool
	Varfile            []string
	TerraformVersion   string
	TerraformEnv       string
	TerraformWorkspace string
	Rules              map[string]*Rule
}

type Rule struct {
	Name    string `hcl:"name,label"`
	Enabled bool   `hcl:"enabled"`
}

func Init() *Config {
	return &Config{
		Debug:              false,
		DeepCheck:          false,
		AwsCredentials:     map[string]string{},
		IgnoreModule:       map[string]bool{},
		IgnoreRule:         map[string]bool{},
		Varfile:            []string{},
		TerraformEnv:       "default",
		TerraformWorkspace: "default",
		Rules:              map[string]*Rule{},
	}
}

func (c *Config) LoadConfig(files ...string) error {
	if b, err := ioutil.ReadFile(".terraform/environment"); err == nil {
		c.TerraformEnv = string(b)
		c.TerraformWorkspace = string(b)
	}

	parser := hclparse.NewParser()
	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			continue
		}

		f, diags := parser.ParseHCLFile(file)
		if diags.HasErrors() {
			return diags
		}

		var raw rawConfig
		diags = gohcl.DecodeBody(f.Body, nil, &raw)
		if diags.HasErrors() {
			return diags
		}

		c.setConfigFromRaw(raw)
	}

	return nil
}

func (c *Config) SetAwsCredentials(accessKey string, secretKey string, profile string, region string) {
	if accessKey != "" {
		c.AwsCredentials["access_key"] = accessKey
	}
	if secretKey != "" {
		c.AwsCredentials["secret_key"] = secretKey
	}
	if profile != "" {
		c.AwsCredentials["profile"] = profile
	}
	if region != "" {
		c.AwsCredentials["region"] = region
	}
}

func (c *Config) HasAwsRegion() bool {
	return c.AwsCredentials["region"] != ""
}

func (c *Config) HasAwsSharedCredentials() bool {
	return c.AwsCredentials["profile"] != "" && c.AwsCredentials["region"] != ""
}

func (c *Config) HasAwsStaticCredentials() bool {
	return c.AwsCredentials["access_key"] != "" && c.AwsCredentials["secret_key"] != "" && c.AwsCredentials["region"] != ""
}

func (c *Config) SetIgnoreModule(ignoreModule string) {
	if ignoreModule == "" {
		return
	}
	ignoreModules := strings.Split(ignoreModule, ",")

	for _, m := range ignoreModules {
		c.IgnoreModule[m] = true
	}
}

func (c *Config) SetIgnoreRule(ignoreRule string) {
	if ignoreRule == "" {
		return
	}
	ignoreRules := strings.Split(ignoreRule, ",")

	for _, r := range ignoreRules {
		c.IgnoreRule[r] = true
	}
}

func (c *Config) SetVarfile(varfile string) {
	// Automatically, `terraform.tfvars` loaded, this priority is the lowest because insert it at the beginning.
	c.Varfile = append([]string{"terraform.tfvars"}, c.Varfile...)

	if varfile == "" {
		return
	}
	varfiles := strings.Split(varfile, ",")
	c.Varfile = append(c.Varfile, varfiles...)
}

func (c *Config) setConfigFromRaw(raw rawConfig) {
	rc := raw.Config

	if rc != nil {
		if rc.DeepCheck != nil {
			c.DeepCheck = *rc.DeepCheck
		}
		if rc.AwsCredentials != nil {
			c.AwsCredentials = *rc.AwsCredentials
		}
		if rc.IgnoreModule != nil {
			c.IgnoreModule = *rc.IgnoreModule
		}
		if rc.IgnoreRule != nil {
			c.IgnoreRule = *rc.IgnoreRule
		}
		if rc.Varfile != nil {
			c.Varfile = *rc.Varfile
		}
		if rc.TerraformVersion != nil {
			c.TerraformVersion = *rc.TerraformVersion
		}
	}

	for _, r := range raw.Rules {
		var rule Rule = r
		c.Rules[rule.Name] = &rule
	}
}
