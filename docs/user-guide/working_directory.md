# Switching working directory with --chdir

The `--chdir` option is available in TFLint like Terraform:

```console
$ tflint --chdir=environments/production
```

Its behavior is the same as [Terraform's behavior](https://developer.hashicorp.com/terraform/cli/commands#switching-working-directory-with-chdir), but there are some TFLint-specific considerations:

- Config file (`.tflint.hcl`) is processed before acting on the `--chdir` option.
- Files specified with relative paths like `--var-file` and `varfile` on config files are resolved against the original working directory.

TFLint also accepts a directory as an argument, but `--chdir` is recommended in most cases. The directory argument is deprecated and may be removed in a future version.
