package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) init(opts Options) int {
	if plugin.IsExperimentalModeEnabled() {
		_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `Experimental mode is enabled. This behavior may change in future versions without notice`)
	}

	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	installed := false
	for _, wd := range workingDirs {
		err := cli.withinChangedDir(wd, func() error {
			cfg, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
			if err != nil {
				if opts.Recursive {
					return fmt.Errorf("Failed to load TFLint config in %s; %w", wd, err)
				} else {
					return fmt.Errorf("Failed to load TFLint config; %w", err)
				}
			}

			for _, pluginCfg := range cfg.Plugins {
				installCfg := plugin.NewInstallConfig(cfg, pluginCfg)

				// If version or source is not set, you need to install it manually
				if installCfg.ManuallyInstalled() {
					continue
				}

				_, err := plugin.FindPluginPath(installCfg)
				if os.IsNotExist(err) {
					if opts.Recursive {
						fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin in %s...\n", pluginCfg.Name, wd)
					} else {
						fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin...\n", pluginCfg.Name)
					}

					_, err = installCfg.Install()
					if err != nil {
						if errors.Is(err, plugin.ErrPluginNotVerified) {
							_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `No signing key configured. Set "signing_key" to verify that the release is signed by the plugin developer`)
						} else {
							return fmt.Errorf("Failed to install a plugin; %w", err)
						}
					}

					installed = true
					fmt.Fprintf(cli.outStream, "Installed \"%s\" (source: %s, version: %s)\n", pluginCfg.Name, pluginCfg.Source, pluginCfg.Version)
				}

				if err != nil {
					if opts.Recursive {
						return fmt.Errorf("Failed to find a plugin in %s; %w", wd, err)
					} else {
						return fmt.Errorf("Failed to find a plugin; %w", err)
					}
				}
			}

			return nil
		})
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, err, map[string][]byte{})
			return ExitCodeError
		}
	}
	if !installed {
		fmt.Fprint(cli.outStream, "All plugins are already installed\n")
	}

	return ExitCodeOK
}
