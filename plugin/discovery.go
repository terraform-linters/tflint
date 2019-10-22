// +build linux darwin

package plugin

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/wata727/tflint/tflint"
)

// PluginRoot is the root directory of the plugins
// This variable is exposed for testing.
var PluginRoot = "~/.tflint.d/plugins"

// Find searches and returns plugins that meet the naming convention.
// All plugins must be placed under `~/.tflint.d/plugins` and
// these must be named `tflint-ruleset-*.so`.
func Find() ([]*Plugin, error) {
	plugins := []*Plugin{}

	pluginDir, err := homedir.Expand(PluginRoot)
	if err != nil {
		return plugins, err
	}
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		log.Printf("[DEBUG] Plugin directory `%s` is not exist", pluginDir)
		return plugins, nil
	}

	log.Printf("[INFO] Finding plugins under `%s`", pluginDir)
	err = filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		matched, err := regexp.Match(`tflint-ruleset-.+\.so`, []byte(info.Name()))
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		p, err := plugin.Open(path)
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return prettyOpenError(err.Error(), info.Name())
		}

		nameSym, err := p.Lookup("Name")
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return fmt.Errorf("Broken plugin `%s` found: The top level `Name` function is undefined", info.Name())
		}
		nameFunc, ok := nameSym.(func() string)
		if !ok {
			return fmt.Errorf("Broken plugin `%s` found: The top level `Name` function must be of type `func() string`", info.Name())
		}

		versionSym, err := p.Lookup("Version")
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return fmt.Errorf("Broken plugin `%s` found: The top level `Version` function is undefined", info.Name())
		}
		versionFunc, ok := versionSym.(func() string)
		if !ok {
			return fmt.Errorf("Broken plugin `%s` found: The top level `Version` function must be of type `func() string`", info.Name())
		}

		rulesSym, err := p.Lookup("NewRules")
		if err != nil {
			log.Printf("[ERROR] %s", err)
			return fmt.Errorf("Broken plugin `%s` found: The top level `NewRules` function is undefined", info.Name())
		}
		rulesFunc, ok := rulesSym.(func() []Rule)
		if !ok {
			return fmt.Errorf("Broken plugin `%s` found: The top level `NewRules` function must be of type `func() []plugin.Rule`", info.Name())
		}

		log.Printf("[INFO] Plugin found: name=%s version=%s", nameFunc(), versionFunc())
		plugins = append(plugins, &Plugin{
			Name:    nameFunc(),
			Version: versionFunc(),
			Rules:   rulesFunc(),
		})
		return nil
	})

	if err != nil {
		return plugins, err
	}
	return plugins, nil
}

func prettyOpenError(message string, name string) error {
	if strings.Contains(message, "plugin was built with a different version of package") {
		return fmt.Errorf("Broken plugin `%s` found: The plugin is built with a different version of TFLint. Should be built with v%s", name, tflint.Version)
	}
	return fmt.Errorf("Broken plugin `%s` found: The plugin is invalid format", name)
}
