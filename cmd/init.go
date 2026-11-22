package cmd

import (
	"errors"
	"fmt"
	"log"
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
			lockfileCreated := false
			if lockfile == nil && lockfilePath != "" {
				lockfile = plugin.NewLockfile(lockfilePath)
				lockfileCreated = true
			}

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

				var locked *plugin.LockedPlugin
				if lockfile != nil {
					locked = lockfile.Plugins[pluginCfg.Source]
				}

				_, err := plugin.FindPluginPath(installCfg)
				if os.IsNotExist(err) {
					if opts.Recursive {
						fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin in %s...\n", pluginCfg.Name, wd)
					} else {
						fmt.Fprintf(cli.outStream, "Installing \"%s\" plugin...\n", pluginCfg.Name)
					}

					var entry *plugin.LockedPlugin
					_, entry, err = installCfg.Install(locked)
					if err != nil {
						if errors.Is(err, plugin.ErrPluginNotVerified) {
							if installCfg.Signature == string(plugin.SignatureModeNone) {
								_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `The plugin signature verification is disabled. Please be aware that disabling verification can pose security risks`)
							} else {
								_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `No signing key or attestations found. The plugin signature is not verified`)
							}
							err = nil
						} else if errors.Is(err, plugin.ErrLegacySigningKeyUsed) {
							_, _ = color.New(color.FgYellow).Fprintln(cli.outStream, `The plugin was signed using a legacy PGP signing key. Please update the plugin to the latest version`)
							err = nil
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
				} else if err == nil && locked != nil && locked.Version != pluginCfg.Version {
					// The plugin is already installed, but the lock entry is for another version.
					// Remove the stale entry because the installed plugin is not verified with the lockfile.
					log.Printf(`[INFO] Remove the stale lock entry for "%s" (locked: %s, declared: %s)`, pluginCfg.Source, locked.Version, pluginCfg.Version)
					delete(lockfile.Plugins, pluginCfg.Source)
				}

				if err != nil {
					if opts.Recursive {
						return fmt.Errorf("Failed to find a plugin in %s; %w", wd, err)
					} else {
						return fmt.Errorf("Failed to find a plugin; %w", err)
					}
				}
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
