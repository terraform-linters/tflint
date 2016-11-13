package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
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

	// TODO: move to other package, and print debug log
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR: Cannot open file %s", filename))
	}
	root, err := parser.Parse(b)
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR: Parse error occurred by %s", filename))
	}
	list, _ := root.Node.(*ast.ObjectList)

	if err := hcl.DecodeObject(c, list.Filter("config").Items[0]); err != nil {
		return err
	}

	return nil
}

func (c *Config) SetIgnoreModule(ignoreModule string) {
	var ignoreModules []string = strings.Split(ignoreModule, ",")

	for _, m := range ignoreModules {
		c.IgnoreModule[m] = true
	}
}

func (c *Config) SetIgnoreRule(ignoreRule string) {
	var ignoreRules []string = strings.Split(ignoreRule, ",")

	for _, r := range ignoreRules {
		c.IgnoreRule[r] = true
	}
}
