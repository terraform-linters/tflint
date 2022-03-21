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
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/host2plugin"
	"github.com/terraform-linters/tflint/tflint"
)

// Discovery searches and launches plugins according the passed configuration.
// If the plugin is not enabled, skip without starting.
func Discovery(config *tflint.Config) (*Plugin, error) {
	clients := map[string]*plugin.Client{}
	rulesets := map[string]*host2plugin.GRPCClient{}

	for _, pluginCfg := range config.Plugins {
		installCfg := NewInstallConfig(config, pluginCfg)
		pluginPath, err := FindPluginPath(installCfg)
		var cmd *exec.Cmd
		if os.IsNotExist(err) {
			if installCfg.ManuallyInstalled() {
				pluginDir, err := getPluginDir(config)
				if err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("Plugin `%s` not found in %s", pluginCfg.Name, pluginDir)
			}
			return nil, fmt.Errorf("Plugin `%s` not found. Did you run `tflint --init`?", pluginCfg.Name)
		} else {
			cmd = exec.Command(pluginPath)
		}

		if pluginCfg.Enabled {
			log.Printf("[INFO] Plugin `%s` found", pluginCfg.Name)

			client := host2plugin.NewClient(&host2plugin.ClientOpts{
				Cmd: cmd,
			})
			rpcClient, err := client.Client()
			if err != nil {
				return nil, pluginClientError(err, pluginCfg)
			}
			raw, err := rpcClient.Dispense("ruleset")
			if err != nil {
				return nil, err
			}
			ruleset := raw.(*host2plugin.GRPCClient)

			clients[pluginCfg.Name] = client
			rulesets[pluginCfg.Name] = ruleset
		} else {
			log.Printf("[INFO] Plugin `%s` found, but the plugin is disabled", pluginCfg.Name)
		}
	}

	return &Plugin{RuleSets: rulesets, clients: clients}, nil
}

// FindPluginPath returns the plugin binary path.
func FindPluginPath(config *InstallConfig) (string, error) {
	dir, err := getPluginDir(config.globalConfig)
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
//   1. `plugin_dir` in a global config
//   2. `TFLINT_PLUGIN_DIR` environment variable
//   3. Current directory (./.tflint.d/plugins)
//   4. Home directory (~/.tflint.d/plugins)
//
// If the environment variable is set, other directories will not be considered,
// but if the current directory does not exist, it will fallback to the home directory.
func getPluginDir(cfg *tflint.Config) (string, error) {
	if cfg.PluginDir != "" {
		return homedir.Expand(cfg.PluginDir)
	}

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
	if runtime.GOOS != "windows" {
		return checkPluginExistance(path)
	}

	returnPath, err := checkPluginExistance(path)
	if os.IsNotExist(err) {
		return checkPluginExistance(path + ".exe")
	}

	return returnPath, err
}

func checkPluginExistance(path string) (string, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", os.ErrNotExist
	} else if !os.IsNotExist(err) && !info.IsDir() {
		return path, nil
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
