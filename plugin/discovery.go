// +build linux darwin

package plugin

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/wata727/tflint/tflint"
)

// Find searches and returns plugins that meet the naming convention.
// All plugins must be placed under `~/.tflint.d/plugins` and
// these must be named `tflint-ruleset-*.so`.
func Find(c *tflint.Config) ([]*Plugin, error) {
	plugins := []*Plugin{}

	pluginDir, err := homedir.Expand(PluginRoot)
	if err != nil {
		return plugins, err
	}

	for _, cfg := range c.Plugins {
		pluginPath := filepath.Join(pluginDir, fmt.Sprintf("tflint-ruleset-%s.so", cfg.Name))
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			return plugins, fmt.Errorf("Plugin `%s` not found in %s", cfg.Name, pluginDir)
		}

		if cfg.Enabled {
			log.Printf("[INFO] Plugin `%s` found", cfg.Name)
			plugin, err := OpenPlugin(pluginPath)
			if err != nil {
				return plugins, err
			}
			plugins = append(plugins, plugin)
		} else {
			log.Printf("[INFO] Plugin `%s` found, but the plugin is disabled", cfg.Name)
		}
	}

	return plugins, nil
}

// OpenPlugin looks up symbol from plugins and intiialize Plugin with functions.
func OpenPlugin(path string) (*Plugin, error) {
	p, err := plugin.Open(path)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, prettyOpenError(err.Error(), path)
	}

	nameSym, err := p.Lookup("Name")
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, fmt.Errorf("Broken plugin `%s` found: The top level `Name` function is undefined", path)
	}
	nameFunc, ok := nameSym.(func() string)
	if !ok {
		return nil, fmt.Errorf("Broken plugin `%s` found: The top level `Name` function must be of type `func() string`", path)
	}

	versionSym, err := p.Lookup("Version")
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, fmt.Errorf("Broken plugin `%s` found: The top level `Version` function is undefined", path)
	}
	versionFunc, ok := versionSym.(func() string)
	if !ok {
		return nil, fmt.Errorf("Broken plugin `%s` found: The top level `Version` function must be of type `func() string`", path)
	}

	rulesSym, err := p.Lookup("NewRules")
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, fmt.Errorf("Broken plugin `%s` found: The top level `NewRules` function is undefined", path)
	}
	rulesFunc, ok := rulesSym.(func() []Rule)
	if !ok {
		return nil, fmt.Errorf("Broken plugin `%s` found: The top level `NewRules` function must be of type `func() []plugin.Rule`", path)
	}

	return &Plugin{
		Name:    nameFunc(),
		Version: versionFunc(),
		Rules:   rulesFunc(),
	}, nil
}

func prettyOpenError(message string, name string) error {
	if strings.Contains(message, "plugin was built with a different version of package") {
		return fmt.Errorf("Broken plugin `%s` found: The plugin is built with a different version of TFLint. Should be built with v%s", name, tflint.Version)
	}
	return fmt.Errorf("Broken plugin `%s` found: The plugin is invalid format", name)
}
