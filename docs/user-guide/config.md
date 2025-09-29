# Configuring TFLint

You can change the behavior not only in CLI flags but also in config files. TFLint loads config files according to the following priority order:

1. File passed by the `--config` option
2. File set by the `TFLINT_CONFIG_FILE` environment variable
3. Current directory `./.tflint.hcl`
4. Current directory `./.tflint.json`
5. Home directory `~/.tflint.hcl`
6. Home directory `~/.tflint.json`

The config file can be written in either [HCL](https://github.com/hashicorp/hcl) or JSON format, determined by the file extension. JSON files use the [HCL-compatible JSON syntax](https://developer.hashicorp.com/terraform/language/syntax/json), following the same structure as Terraform's `.tf.json` files. An HCL example is shown below:

```hcl
tflint {
  required_version = ">= 0.50"
}

config {
  format = "compact"
  plugin_dir = "~/.tflint.d/plugins"

  call_module_type = "local"
  force = false
  disabled_by_default = false

  ignore_module = {
    "terraform-aws-modules/vpc/aws"            = true
    "terraform-aws-modules/security-group/aws" = true
  }

  varfile = ["example1.tfvars", "example2.tfvars"]
  variables = ["foo=bar", "bar=[\"baz\"]"]
}

plugin "aws" {
  enabled = true
  version = "0.4.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

rule "aws_instance_invalid_type" {
  enabled = false
}
```

The same configuration can be written in JSON format as `.tflint.json`:

```json
{
  "tflint": {
    "required_version": ">= 0.50"
  },
  "config": {
    "format": "compact",
    "plugin_dir": "~/.tflint.d/plugins",
    "call_module_type": "local",
    "force": false,
    "disabled_by_default": false,
    "ignore_module": {
      "terraform-aws-modules/vpc/aws": true,
      "terraform-aws-modules/security-group/aws": true
    },
    "varfile": ["example1.tfvars", "example2.tfvars"],
    "variables": ["foo=bar", "bar=[\"baz\"]"]
  },
  "plugin": {
    "aws": {
      "enabled": true,
      "version": "0.4.0",
      "source": "github.com/terraform-linters/tflint-ruleset-aws"
    }
  },
  "rule": {
    "aws_instance_invalid_type": {
      "enabled": false
    }
  }
}
```

The file path is resolved relative to the module directory when `--chdir` or `--recursive` is used. To use a config file from the working directory when recursing, pass an absolute path:

```sh
tflint --recursive --config "$(pwd)/.tflint.hcl"
# or
tflint --recursive --config "$(pwd)/.tflint.json"
```

### `required_version`

Restrict the TFLint version used. This is almost the same as [Terraform's `required_version`](https://developer.hashicorp.com/terraform/language/settings#specifying-a-required-terraform-version).
You can write version constraints in the same way.

### `format`

CLI flag: `--format`

Change the output format. The following values are valid:

- default
- json
- checkstyle
- junit
- compact
- sarif

In recursive mode (`--recursive`), this field will be ignored in configuration files and must be set via a flag.

### `plugin_dir`

Set the plugin directory. The default is `~/.tflint.d/plugins` (or `./.tflint.d/plugins`). See also [Configuring Plugins](plugins.md#advanced-usage)

### `call_module_type`

CLI flag: `--call-module-type`

Select types of module to call. The following values are valid:

- all
- local (default)
- none

If you select `all`, you can call all (local and remote) modules. See [Calling Modules](./calling-modules.md).

```hcl
config {
  call_module_type = "all"
}
```

```console
$ tflint --call-module-type=all
```

### `force`

CLI flag: `--force`

Return zero exit status even if issues found. TFLint returns the following exit statuses on exit by default:

- 0: No issues found
- 1: Errors occurred
- 2: No errors occurred, but issues found

In recursive mode (`--recursive`), this field will be ignored in configuration files and must be set via a flag.

### `disabled_by_default`

CLI flag: `--only`

Only enable rules specifically enabled in the config or on the command line. All other rules, including defaults, are disabled. Note, usage of `--only` on the command line will ignore other rules passed in via `--enable-rule` or `--disable-rule`.

```hcl
config {
  disabled_by_default = true
  # other options here...
}

rule "aws_instance_invalid_type" {
  enabled = true
}

rule "aws_instance_previous_type" {
  enabled = true
}
```

```console
$ tflint --only aws_instance_invalid_type --only aws_instance_previous_type
```

### `ignore_module`

CLI flag: `--ignore-module`

Adding a module source to `ignore_module` will cause it to be ignored when [calling modules](./calling-modules.md). Note that you need to specify module sources rather than module ids for backward compatibility.

```hcl
config {
  call_module_type = "all"
  ignore_module = {
    "terraform-aws-modules/vpc/aws"            = true
    "terraform-aws-modules/security-group/aws" = true
  }
}
```

```console
$ tflint --ignore-module terraform-aws-modules/vpc/aws --ignore-module terraform-aws-modules/security-group/aws
```

### `varfile`

CLI flag: `--var-file`

Set Terraform variables from `tfvars` files. If `terraform.tfvars` or any `*.auto.tfvars` files are present, they will be automatically loaded.

```hcl
config {
  varfile = ["example1.tfvars", "example2.tfvars"]
}
```

```console
$ tflint --var-file example1.tfvars --var-file example2.tfvars
```

### `variables`

CLI flag: `--var`

Set a Terraform variable from a passed value. This flag can be set multiple times.

```hcl
config {
  variables = ["foo=bar", "bar=[\"baz\"]"]
}
```

```console
$ tflint --var "foo=bar" --var "bar=[\"baz\"]"
```

### `rule` blocks

CLI flag: `--enable-rule`, `--disable-rule`

You can configure TFLint rules using `rule` blocks. Each rule's implementation specifies whether it will be enabled by default. In some rulesets, the majority of rules are disabled by default. Use `rule` blocks to enable them:

```hcl
rule "terraform_unused_declarations" {
  enabled = true
}
```

The `enabled` attribute is required for all `rule` blocks. For rules that are enabled by default, set `enabled = false` to disable the rule:

```hcl
rule "aws_instance_previous_type" {
  enabled = false
}
```

Some rules support additional attributes that configure their behavior. See the documentation for each rule for details.

### `plugin` blocks

You can declare the plugin to use. See [Configuring Plugins](plugins.md)

## Rule config priority

The priority of rule configs is as follows:

1. `--only` (CLI flag)
2. `--enable-rule`, `--disable-rule` (CLI flag)
3. `rule` blocks (config file)
4. `preset` (config file, tflint-ruleset-terraform only)
5. `disabled_by_default` (config file)
