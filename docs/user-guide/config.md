# Configuring TFLint

You can change the behavior not only in CLI flags but also in config files. By default, TFLint looks up `.tflint.hcl` according to the following priority:

- Current directory (`./.tflint.hcl`)
- Home directory (`~/.tflint.hcl`)

The config file is written in [HCL](https://github.com/hashicorp/hcl). An example is shown below:

```hcl
config {
  plugin_dir = "~/.tflint.d/plugins"

  module = true
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

You can also use another file as a config file with the `--config` option:

```
$ tflint --config other_config.hcl
```

### `plugin_dir`

Set the plugin directory. The default is `~/.tflint.d/plugins` (or `./.tflint.d/plugins`). See also [Configuring Plugins](plugins.md#advanced-usage)

### `module`

CLI flag: `--module`

Enable [Module Inspection](module-inspection.md).

### `force`

CLI flag: `--force`

Return zero exit status even if issues found. TFLint returns the following exit statuses on exit by default:

- 0: No issues found
- 1: Errors occurred
- 2: No errors occurred, but issues found

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

Skip inspections for module calls in [Module Inspection](module-inspection.md). Note that you need to specify module sources rather than module ids for backward compatibility.

```hcl
config {
  module = true
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
