package config

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/wata727/tflint/loader"
)

type Config struct {
	Debug          bool
	DeepCheck      bool              `hcl:"deep_check"`
	AwsCredentials map[string]string `hcl:"aws_credentials"`
	IgnoreModule   map[string]bool   `hcl:"ignore_module"`
	IgnoreRule     map[string]bool   `hcl:"ignore_rule"`
}

func Init() *Config {
	return &Config{
		Debug:          false,
		DeepCheck:      false,
		AwsCredentials: map[string]string{},
		IgnoreModule:   map[string]bool{},
		IgnoreRule:     map[string]bool{},
	}
}

func (c *Config) LoadConfig(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return nil
	}

	l := loader.NewLoader(c.Debug)
	if err := l.LoadTemplate(filename); err != nil {
		return nil
	}

	if err := hcl.DecodeObject(c, l.ListMap[filename].Filter("config").Items[0]); err != nil {
		return err
	}

	return nil
}

func (c *Config) SetAwsCredentials(accessKey string, secretKey string, region string) {
	if accessKey != "" {
		c.AwsCredentials["access_key"] = accessKey
	}
	if secretKey != "" {
		c.AwsCredentials["secret_key"] = secretKey
	}
	if region != "" {
		c.AwsCredentials["region"] = region
	}
}

func (c *Config) HasAwsRegion() bool {
	return c.AwsCredentials["region"] != ""
}

func (c *Config) HasAwsCredentials() bool {
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
