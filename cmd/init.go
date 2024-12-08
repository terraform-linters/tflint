package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/plugin"
	"github.com/terraform-linters/tflint/tflint"
)

func (cli *CLI) init(opts Options) int {
	workingDirs, err := findWorkingDirs(opts)
	if err != nil {
		cli.formatter.Print(tflint.Issues{}, fmt.Errorf("Failed to find workspaces; %w", err), map[string][]byte{})
		return ExitCodeError
	}

	var builder strings.Builder

	if opts.Recursive {
		fmt.Fprint(cli.outStream, "Installing plugins on each working directory...\n\n")
	}

	any_installed := false
	for _, wd := range workingDirs {
		builder.Reset()
		err := cli.withinChangedDir(wd, func() error {
			installed := false
			if opts.Recursive {
				builder.WriteString("====================================================\n")
				builder.WriteString(fmt.Sprintf("working directory: %s\n\n", wd))
			}

			cfg, err := tflint.LoadConfig(afero.Afero{Fs: afero.NewOsFs()}, opts.Config)
			if err != nil {
				fmt.Fprint(cli.outStream, builder.String())
				return fmt.Errorf("Failed to load TFLint config; %w", err)
			}

			found := false
			for _, pluginCfg := range cfg.Plugins {
				installCfg := plugin.NewInstallConfig(cfg, pluginCfg)

				// If version or source is not set, you need to install it manually
				if installCfg.ManuallyInstalled() {
					continue
				}
				found = true

				_, err := plugin.FindPluginPath(installCfg)
				if os.IsNotExist(err) {
					fmt.Fprint(cli.outStream, builder.String())
					builder.Reset()
					fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin...\n", pluginCfg.Name)

					sigchecker := plugin.NewSignatureChecker(installCfg)
					if !sigchecker.HasSigningKey() {
						_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `No signing key configured. Set "signing_key" to verify that the release is signed by the plugin developer`)
					}

					_, err = installCfg.Install()
					if err != nil {
						return fmt.Errorf("Failed to install a plugin; %w", err)
					}

					any_installed = true
					installed = true
					fmt.Fprintf(cli.outStream, "Installed \"%s\" (source: %s, version: %s)\n", pluginCfg.Name, pluginCfg.Source, pluginCfg.Version)
				}

				if err != nil {
					fmt.Fprint(cli.outStream, builder.String())
					return fmt.Errorf("Failed to find a plugin; %w", err)
				}

				builder.WriteString(fmt.Sprintf("Plugin \"%s\" is already installed\n", pluginCfg.Name))
			}

			if opts.Recursive && !found {
				builder.WriteString("No plugins to install\n")
			}

			if installed || !opts.Recursive {
				fmt.Fprint(cli.outStream, builder.String())
				return nil
			}

			// If there are no changes, send logs to debug
			prefix := "[DEBUG]   "
			lines := strings.Split(builder.String(), "\n")

			for _, line := range lines {
				log.Printf("%s%s", prefix, line)
			}

			return nil
		})
		if err != nil {
			cli.formatter.Print(tflint.Issues{}, err, map[string][]byte{})
			return ExitCodeError
		}
	}
	if opts.Recursive && !any_installed {
		fmt.Fprint(cli.outStream, "All plugins are already installed\n")
	}

	return ExitCodeOK
}
