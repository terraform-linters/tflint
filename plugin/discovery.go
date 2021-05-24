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

// Discovery searches and launches plugins according the passed configuration.
// If the plugin is not enabled, skip without starting.
// The AWS plugin is treated specially. Plugins for which no version is specified will launch the bundled plugin
// instead of returning an error. This is a process for backward compatibility.
func Discovery(config *tflint.Config) (*Plugin, error) {
	clients := map[string]*plugin.Client{}
	rulesets := map[string]*tfplugin.Client{}

	for _, cfg := range config.Plugins {
		installCfg := NewInstallConfig(cfg)
		pluginPath, err := FindPluginPath(installCfg)
		var cmd *exec.Cmd
		if os.IsNotExist(err) {
			if cfg.Name == "aws" && installCfg.ManuallyInstalled() {
				log.Print("[INFO] Plugin `aws` is not installed, but bundled plugins are available.")
				self, err := os.Executable()
				if err != nil {
					return nil, err
				}
				cmd = exec.Command(self, "--act-as-aws-plugin")
			} else {
				if installCfg.ManuallyInstalled() {
					pluginDir, err := getPluginDir()
					if err != nil {
						return nil, err
					}
					return nil, fmt.Errorf("Plugin `%s` not found in %s", cfg.Name, pluginDir)
				}
				return nil, fmt.Errorf("Plugin `%s` not found. Did you run `tflint --init`?", cfg.Name)
			}
		} else {
			cmd = exec.Command(pluginPath)
		}

		if cfg.Enabled {
			log.Printf("[INFO] Plugin `%s` found", cfg.Name)

			client := tfplugin.NewClient(&tfplugin.ClientOpts{
				Cmd: cmd,
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

// FindPluginPath returns the plugin binary path.
func FindPluginPath(config *InstallConfig) (string, error) {
	dir, err := getPluginDir()
	if err != nil {
		return "", err
	}

	path, err := findPluginPath(filepath.Join(dir, config.InstallPath()))
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Find plugin path: %s", path)
	return path, err
}

// getPluginDir returns the base plugin directory.
// Adopted with the following priorities:
//
//   1. `TFLINT_PLUGIN_DIR` environment variable
//   2. Current directory (./.tflint.d/plugins)
//   3. Home directory (~/.tflint.d/plugins)
//
// If the environment variable is set, other directories will not be considered,
// but if the current directory does not exist, it will fallback to the home directory.
func getPluginDir() (string, error) {
	if dir := os.Getenv("TFLINT_PLUGIN_DIR"); dir != "" {
		return dir, nil
	}

	_, err := os.Stat(localPluginRoot)
	if os.IsNotExist(err) {
		return homedir.Expand(PluginRoot)
	}

	return localPluginRoot, err
}

// findPluginPath returns the path of the existing plugin.
// Only in the case of Windows, the pattern with the `.exe` is also considered,
// and if it exists, the extension is added to the argument.
func findPluginPath(path string) (string, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) && runtime.GOOS != "windows" {
		return "", os.ErrNotExist
	} else if !os.IsNotExist(err) {
		return path, nil
	}

	if _, err := os.Stat(path + ".exe"); !os.IsNotExist(err) {
		return path + ".exe", nil
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
