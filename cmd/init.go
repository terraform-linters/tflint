package cmd

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

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

			lockfilePath := plugin.LockfilePath(cfg.ConfigPath())
			lockfile, err := plugin.LoadLockfile(lockfilePath)
			if err != nil {
				return err
			}

			// Collect plugins to be installed automatically and check if they are already installed
			names := []string{}
			alreadyInstalled := map[string]bool{}
			// Plugins declared in the config. Used to remove unused entries from the lockfile.
			declared := map[string]string{}
			for _, name := range slices.Sorted(maps.Keys(cfg.Plugins)) {
				pluginCfg := cfg.Plugins[name]
				installCfg := plugin.NewInstallConfig(cfg, pluginCfg)

				// If version or source is not set, you need to install it manually
				if installCfg.ManuallyInstalled() {
					continue
				}

				if version, exists := declared[pluginCfg.Source]; exists && version != pluginCfg.Version {
					return fmt.Errorf(`Failed to install a plugin; The plugin source "%s" is declared with multiple versions (%s, %s)`, pluginCfg.Source, version, pluginCfg.Version)
				}
				declared[pluginCfg.Source] = pluginCfg.Version

				_, err := plugin.FindPluginPath(installCfg)
				if err != nil && !os.IsNotExist(err) {
					if opts.Recursive {
						return fmt.Errorf("Failed to find a plugin in %s; %w", wd, err)
					} else {
						return fmt.Errorf("Failed to find a plugin; %w", err)
					}
				}
				names = append(names, name)
				alreadyInstalled[name] = err == nil
			}

			// The lockfile is created only when installing a new plugin, so as not to affect
			// existing users who have already installed plugins.
			lockfileCreated := false
			if lockfile == nil && lockfilePath != "" && slices.ContainsFunc(names, func(name string) bool { return !alreadyInstalled[name] }) {
				lockfile = plugin.NewLockfile(lockfilePath)
				lockfileCreated = true
			}

			for _, name := range names {
				pluginCfg := cfg.Plugins[name]
				installCfg := plugin.NewInstallConfig(cfg, pluginCfg)

				var locked *plugin.LockedPlugin
				if lockfile != nil {
					locked = lockfile.Plugins[pluginCfg.Source]
				}

				if alreadyInstalled[name] {
					// If the lockfile is used, all plugins must be recorded with the declared version.
					// Reinstall plugins that are installed but not recorded in the lockfile.
					if lockfile == nil || (locked != nil && locked.Version == pluginCfg.Version) {
						continue
					}
					if opts.Recursive {
						fmt.Fprintf(cli.outStream, "Reinstalling \"%s\" plugin in %s to record it in the lockfile...\n", pluginCfg.Name, wd)
					} else {
						fmt.Fprintf(cli.outStream, "Reinstalling \"%s\" plugin to record it in the lockfile...\n", pluginCfg.Name)
					}
				} else {
					if opts.Recursive {
						fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin in %s...\n", pluginCfg.Name, wd)
					} else {
						fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin...\n", pluginCfg.Name)
					}
				}

				_, entry, err := installCfg.Install(locked)
				if err != nil {
					if errors.Is(err, plugin.ErrPluginNotVerified) {
						if installCfg.Signature == string(plugin.SignatureModeNone) {
							_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `The plugin signature verification is disabled. Please be aware that disabling verification can pose security risks`)
						} else {
							_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `No signing key or attestations found. The plugin signature is not verified`)
						}
					} else if errors.Is(err, plugin.ErrLegacySigningKeyUsed) {
						_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `The plugin was signed using a legacy PGP signing key. Please update the plugin to the latest version`)
					} else {
						// Save the lock entries of plugins installed so far
						if saveErr := cli.saveLockfile(lockfile, lockfileCreated); saveErr != nil {
							return saveErr
						}
						return fmt.Errorf("Failed to install a plugin; %w", err)
					}
				}

				if lockfile != nil {
					lockfile.Plugins[pluginCfg.Source] = entry
				}
				installed = true
				fmt.Fprintf(cli.outStream, "Installed \"%s\" (source: %s, version: %s)\n", pluginCfg.Name, pluginCfg.Source, pluginCfg.Version)
			}

			// Remove entries for plugins that are no longer declared
			if lockfile != nil {
				for source := range lockfile.Plugins {
					if _, exists := declared[source]; !exists {
						delete(lockfile.Plugins, source)
					}
				}
			}

			return cli.saveLockfile(lockfile, lockfileCreated)
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

// saveLockfile writes the lockfile if it has been changed.
// A new lockfile is not created if there are no entries.
func (cli *CLI) saveLockfile(lockfile *plugin.Lockfile, created bool) error {
	if lockfile == nil {
		return nil
	}
	if created && len(lockfile.Plugins) == 0 {
		return nil
	}

	written, err := lockfile.Save()
	if err != nil {
		return err
	}
	if written {
		if created {
			fmt.Fprintf(cli.outStream, "Created lockfile %s\n", lockfile.Path())
		} else {
			fmt.Fprintf(cli.outStream, "Updated lockfile %s\n", lockfile.Path())
		}
	}
	return nil
}
