# Configuring TFLint

You can change the behavior not only in CLI flags but also in configuration files. By default, TFLint looks up `.tflint.hcl` according to the following priority:

- Current directory (`./.tflint.hcl`)
- Home directory (`~/.tflint.hcl`)

The config file is written in [HCL](https://github.com/hashicorp/hcl/tree/hcl2). An example is shown below:

```hcl
config {
  module = true
  deep_check = true
  force = false

  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }

  ignore_module = {
    "github.com/terraform-linters/example-module" = true
  }

  varfile = ["example1.tfvars", "example2.tfvars"]

  variables = ["foo=bar", "bar=[\"baz\"]"]

  tags = []"foo", "bar"]
}

rule "aws_instance_invalid_type" {
  enabled = false
}

rule "aws_instance_previous_type" {
  enabled = false
}

plugin "example" {
  enabled = true
}
```

You can also use another file as a config file with the `--config` option:

```
$ tflint --config other_config.hcl
```

## `module`

CLI flag: `--module`

Enable [Module inspection](advanced.md#module-inspection).

## `deep_check`

CLI flag: `--deep`

Enable [Deep checking](advanced.md#deep-checking).

## `force`

CLI flag: `--force`

Return zero exit status even if issues found. TFLint returns non-zero exit status by default. See [Exit statuses](../../README.md#exit-statuses).

## `aws_credentials`

CLI flag: `--aws-access-key`, `--aws-secret-key`, `--aws-profile`, `--aws-creds-file` and `--aws-region`

Configure AWS service crendetials. See [Credentials](credentials.md).

## `ignore_module`

CLI flag: `--ignore-module`

Skip inspections for the specified comma-separated module calls. Note that you need to pass module sources rather than module ids for backward compatibility. See [Module inspection](advanced.md#module-inspection).

## `varfile`

CLI flag: `--var-file`

Set Terraform variables from `tfvars` files. If `terraform.tfvars` or any `*.auto.tfvars` files are present, they will be automatically loaded.

## `variables`

CLI flag: `--var`

Set a Terraform variable from a passed value. This flag can be set multiple times.

## `tags`

CLI flag: `--tag`

Check that AWS resources have the expected tag keys. This flag can be set multiple times.

## `rule` blocks

CLI flag: `--enable-rule`, `--disable-rule`

You can make settings for each rule in the `rule` block. All rules have the `enabled` attribute, and when it is false, the rule is ignored from inspection.

```hcl
rule "aws_instance_previous_type" {
  enabled = false
}
```

Each rule can have its own configs. See the documentation for each rule for details.

## `plugin` blocks

You can enable each plugin in the `plugin` block. Currently, it can set only `enabled` option. See [Extending TFLint](extend.md) for details.

```
plugin "example" {
  enabled = true
}
```
