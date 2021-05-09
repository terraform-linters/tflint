package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	tfplugin "github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) init(opts Options) int {
	cfg, err := tflint.LoadConfig(opts.Config)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to load TFLint config", err), map[string][]byte{})
		return ExitCodeError
	}

	for _, pluginCfg := range cfg.Plugins {
		installCfg := tfplugin.NewInstallConfig(pluginCfg)

		// If version or source is not set, you need to install it manually
		if installCfg.ManuallyInstalled() {
			continue
		}

		_, err := tfplugin.FindPluginPath(installCfg)
		if os.IsNotExist(err) {
			fmt.Fprintf(cli.outStream, "Installing `%s` plugin...\n", pluginCfg.Name)

			sigchecker := tfplugin.NewSignatureChecker(installCfg)
			if !sigchecker.HasSigningKey() {
				color.New(color.FgYellow).Fprintln(cli.outStream, "No signing key configured. Set `signing_key` to verify that the release is signed by the plugin developer")
			}

			_, err = installCfg.Install()
			if err != nil {
				cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to install a plugin", err), map[string][]byte{})
				return ExitCodeError
			}

			fmt.Fprintf(cli.outStream, "Installed `%s` (source: %s, version: %s)\n", pluginCfg.Name, pluginCfg.Source, pluginCfg.Version)
			continue
		}

		if err != nil {
			cli.formatter.Print(tflint.Issues{}, tflint.NewContextError("Failed to find a plugin", err), map[string][]byte{})
			return ExitCodeError
		}

		fmt.Fprintf(cli.outStream, "Plugin `%s` is already installed\n", pluginCfg.Name)
	}

	return ExitCodeOK
}
