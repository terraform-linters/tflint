package config

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/wata727/tflint/loader"
)

type Config struct {
	Debug        bool
	IgnoreModule map[string]bool `hcl:"ignore_module"`
	IgnoreRule   map[string]bool `hcl:"ignore_rule"`
}

func Init() *Config {
	return &Config{
		Debug:        false,
		IgnoreModule: map[string]bool{},
		IgnoreRule:   map[string]bool{},
	}
}

func (c *Config) LoadConfig(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return nil
	}

	l := loader.NewLoader(c.Debug)
	if err := l.LoadFile(filename); err != nil {
		return nil
	}

	if err := hcl.DecodeObject(c, l.ListMap[filename].Filter("config").Items[0]); err != nil {
		return err
	}

	return nil
}

func (c *Config) SetIgnoreModule(ignoreModule string) {
	ignoreModules := strings.Split(ignoreModule, ",")

	for _, m := range ignoreModules {
		c.IgnoreModule[m] = true
	}
}

func (c *Config) SetIgnoreRule(ignoreRule string) {
	ignoreRules := strings.Split(ignoreRule, ",")

	for _, r := range ignoreRules {
		c.IgnoreRule[r] = true
	}
}
