package plugin

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/mitchellh/go-homedir"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

// Discovery searches plugins according the passed configuration
// The search priority of plugins is as follows:
//
//   1. Current directory (./.tflint.d/plugins)
//   2. Home directory (~/.tflint.d/plugins)
//
// Files under these directories that satisfy the "tflint-ruleset-*" naming rules
// enabled in the configuration are treated as plugins.
//
// If the `TFLINT_PLUGIN_DIR` environment variable is set, ignore the above and refer to that directory.
func Discovery(config *tflint.Config) (*Plugin, error) {
	if dir := os.Getenv("TFLINT_PLUGIN_DIR"); dir != "" {
		return findPlugins(config, dir)
	}

	if _, err := os.Stat(localPluginRoot); !os.IsNotExist(err) {
		return findPlugins(config, localPluginRoot)
	}

	pluginDir, err := homedir.Expand(PluginRoot)
	if err != nil {
		return nil, err
	}
	return findPlugins(config, pluginDir)
}

func findPlugins(config *tflint.Config, dir string) (*Plugin, error) {
	clients := map[string]*plugin.Client{}
	rulesets := map[string]*tfplugin.Client{}

	for _, cfg := range config.Plugins {
		pluginPath, err := getPluginPath(dir, cfg.Name)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Plugin `%s` not found in %s", cfg.Name, dir)
		}

		if cfg.Enabled {
			log.Printf("[INFO] Plugin `%s` found", cfg.Name)

			client := tfplugin.NewClient(&tfplugin.ClientOpts{
				Cmd: exec.Command(pluginPath),
			})
			rpcClient, err := client.Client()
			if err != nil {
				return nil, pluginClientError(err, cfg)
			}
			raw, err := rpcClient.Dispense("ruleset")
			if err != nil {
				return nil, err
			}
			ruleset := raw.(*tfplugin.Client)

			clients[cfg.Name] = client
			rulesets[cfg.Name] = ruleset
		} else {
			log.Printf("[INFO] Plugin `%s` found, but the plugin is disabled", cfg.Name)
		}
	}

	return &Plugin{RuleSets: rulesets, clients: clients}, nil
}

func getPluginPath(dir string, name string) (string, error) {
	pluginPath := filepath.Join(dir, fmt.Sprintf("tflint-ruleset-%s", name))

	_, err := os.Stat(pluginPath)
	if os.IsNotExist(err) && runtime.GOOS != "windows" {
		return "", os.ErrNotExist
	} else if !os.IsNotExist(err) {
		return pluginPath, nil
	}

	if _, err := os.Stat(pluginPath + ".exe"); !os.IsNotExist(err) {
		return pluginPath + ".exe", nil
	}

	return "", os.ErrNotExist
}

func pluginClientError(err error, config *tflint.PluginConfig) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "Incompatible API version") {
		message := err.Error()
		message = strings.Replace(
			message,
			"Incompatible API version with plugin.",
			fmt.Sprintf(`Incompatible API version with plugin "%s".`, config.Name),
			-1,
		)
		message = strings.Replace(message, "Client versions:", "TFLint versions:", -1)

		return errors.New(message)
	}

	return err
}
