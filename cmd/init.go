package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) init(opts Options) int {
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to load TFLint config; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	for _, pluginCfg := range cfg.Plugins {
		installCfg := plugin.NewInstallConfig(cfg, pluginCfg)

		// If version or source is not set, you need to install it manually
		if installCfg.ManuallyInstalled() {
			continue
		}

		_, err := plugin.FindPluginPath(installCfg)
		if os.IsNotExist(err) {
			fmt.Fprintf(cli.outStream, "Installing `%s` plugin...\n", pluginCfg.Name)

			sigchecker := plugin.NewSignatureChecker(installCfg)
			if !sigchecker.HasSigningKey() {
				_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, "No signing key configured. Set `signing_key` to verify that the release is signed by the plugin developer")
			}

			_, err = installCfg.Install()
			if err != nil {
				cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to install a plugin; %w", err), map[string][]byte{})
				return ExitCodeError
			}

			fmt.Fprintf(cli.outStream, "Installed `%s` (source: %s, version: %s)\n", pluginCfg.Name, pluginCfg.Source, pluginCfg.Version)
			continue
		}

		if err != nil {
			cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find a plugin; %w", err), map[string][]byte{})
			return ExitCodeError
		}

		fmt.Fprintf(cli.outStream, "Plugin `%s` is already installed\n", pluginCfg.Name)
	}

	return ExitCodeOK
}
