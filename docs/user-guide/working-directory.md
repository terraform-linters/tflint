# Switching working directory

TFLint has `--chdir` and `--recursive` flags to inspect modules that are different from the current directory.

The `--chdir` flag is available just like Terraform:

```console
$ tflint --chdir=environments/production
```

Its behavior is the same as [Terraform's behavior](https://developer.hashicorp.com/terraform/cli/commands#switching-working-directory-with-chdir). You should be aware of the following points:

- Config files are loaded after acting on the `--chdir` option.
  - This means that `tflint --chdir=dir` will loads `dir/.tflint.hcl` instead of `./.tflint.hcl`.
- Relative paths are always resolved against the changed directory.
  - If you want to refer to the file in the original working directory, it is recommended to pass the absolute path using realpath(1) etc. e.g. `tflint --config=$(realpath .tflint.hcl)`.
- The `path.cwd` represents the original working directory. This is the same behavior as using `--chdir` in Terraform.

TFLint also accepts a directory as an argument, but `--chdir` is recommended in most cases. The directory argument is deprecated and may be removed in a future version.

The `--recursive` flag enables recursive inspection. This is the same as running with `--chdir` for each directory.

```console
$ tflint --recursive
```

It takes no arguments in recursive mode. Passing a directory or file name will result in an error.

These flags are also valid for `--init` and `--version`. Recursive init is required when installing required plugins all at once:

```console
$ tflint --recursive --init
$ tflint --recursive --version
$ tflint --recursive
```
