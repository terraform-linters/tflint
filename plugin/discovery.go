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
	"github.com/terraform-linters/tflint/internal/workingdir"
	"github.com/terraform-linters/tflint/tflint"
)

// Discovery searches and launches plugins according the passed configuration.
// If the plugin is not enabled, skip without starting.
// The Terraform Language plugin is treated specially. Plugins for which no version
// is specified will launch the bundled plugin instead of returning an error.
func Discovery(config *tflint.Config) (*Plugin, error) {
	return DiscoveryAt(config, "")
}

// DiscoveryAt resolves plugin paths and starts plugins in dir. It does not
// change the host's working directory.
func DiscoveryAt(config *tflint.Config, dir string) (_ *Plugin, err error) {
	clients := map[string]*plugin.Client{}
	rulesets := map[string]*host2plugin.Client{}
	plugins := &Plugin{RuleSets: rulesets, clients: clients}
	defer func() {
		if err != nil {
			plugins.Clean()
		}
	}()

	for _, pluginCfg := range config.Plugins {
		installCfg := NewInstallConfig(config, pluginCfg)
		pluginPath, err := findPluginPathAt(installCfg, dir)
		var cmd *exec.Cmd
		if os.IsNotExist(err) {
			if pluginCfg.Name == "terraform" && installCfg.ManuallyInstalled() {
				log.Print(`[INFO] Plugin "terraform" is not installed, but the bundled plugin is available.`)
				self, err := os.Executable()
				if err != nil {
					return nil, err
				}
				cmd = exec.Command(self, "--act-as-bundled-plugin")
			} else {
				if installCfg.ManuallyInstalled() {
					pluginDir, err := getPluginDirAt(config, dir)
					if err != nil {
						return nil, err
					}
					return nil, fmt.Errorf(`Plugin "%s" not found in %s`, pluginCfg.Name, pluginDir)
				}
				return nil, fmt.Errorf(`Plugin "%s" not found. Did you run "tflint --init"?`, pluginCfg.Name)
			}
		} else {
			cmd = exec.Command(pluginPath)
		}
		cmd.Dir = dir

		if pluginCfg.Enabled {
			log.Printf(`[INFO] Plugin "%s" found`, pluginCfg.Name)

			client := host2plugin.NewClient(&host2plugin.ClientOpts{
				Cmd: cmd,
			})
			clients[pluginCfg.Name] = client
			rpcClient, err := client.Client()
			if err != nil {
				return nil, pluginClientError(err, pluginCfg)
			}
			raw, err := rpcClient.Dispense("ruleset")
			if err != nil {
				return nil, err
			}
			ruleset := raw.(*host2plugin.Client)

			rulesets[pluginCfg.Name] = ruleset
		} else {
			log.Printf(`[INFO] Plugin "%s" found, but the plugin is disabled`, pluginCfg.Name)
		}
	}

	return plugins, nil
}

// FindPluginPath returns the plugin binary path.
func FindPluginPath(config *InstallConfig) (string, error) {
	return findPluginPathAt(config, "")
}

func findPluginPathAt(config *InstallConfig, workingDir string) (string, error) {
	dir, err := getPluginDirAt(config.globalConfig, workingDir)
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
//  1. `plugin_dir` in a global config
//  2. `TFLINT_PLUGIN_DIR` environment variable
//  3. Current directory (./.tflint.d/plugins)
//  4. Home directory (~/.tflint.d/plugins)
//
// If the environment variable is set, other directories will not be considered,
// but if the current directory does not exist, it will fallback to the home directory.
func getPluginDir(cfg *tflint.Config) (string, error) {
	return getPluginDirAt(cfg, "")
}

func getPluginDirAt(cfg *tflint.Config, dir string) (string, error) {
	if cfg.PluginDir != "" {
		path, err := homedir.Expand(cfg.PluginDir)
		return workingdir.Resolve(dir, path), err
	}

	if path := os.Getenv("TFLINT_PLUGIN_DIR"); path != "" {
		return workingdir.Resolve(dir, path), nil
	}

	localRoot := workingdir.Resolve(dir, localPluginRoot)
	_, err := os.Stat(localRoot)
	if os.IsNotExist(err) {
		return homedir.Expand(PluginRoot)
	}

	return localRoot, err
}

// findPluginPath returns the path of the existing plugin.
// Only in the case of Windows, the pattern with the `.exe` is also considered,
// and if it exists, the extension is added to the argument.
func findPluginPath(path string) (string, error) {
	if runtime.GOOS != "windows" {
		return checkPluginExistence(path)
	}

	returnPath, err := checkPluginExistence(path)
	if os.IsNotExist(err) {
		return checkPluginExistence(path + ".exe")
	}

	return returnPath, err
}

func checkPluginExistence(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	// directory is invalid as a plugin path
	if info.IsDir() {
		return "", os.ErrNotExist
	}

	return path, nil
}

func pluginClientError(err error, config *tflint.PluginConfig) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	search := "Incompatible API version with plugin."

	if strings.Contains(message, search) {
		message = strings.ReplaceAll(
			message,
			search,
			fmt.Sprintf(`TFLint is not compatible with this version of the %q plugin. A newer TFLint or plugin version may be required.`, config.Name),
		)

		return errors.New(message)
	}

	return err
}
