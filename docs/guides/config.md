# Configuring TFLint

You can change the behavior not only in CLI flags but also in configuration files. By default, TFLint looks up `.tflint.hcl` according to the following priority:

- Current directory (`./.tflint.hcl`)
- Home directory (`~/.tflint.hcl`)

The config file is written in [HCL](https://github.com/hashicorp/hcl2). An example is shown below:

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
    "github.com/wata727/example-module" = true
  }

  varfile = ["example1.tfvars", "example2.tfvars"]

  variables = ["foo=bar", "bar=[\"baz\"]"]
}

rule "aws_instance_invalid_type" {
  enabled = false
}

rule "aws_instance_previous_type" {
  enabled = false
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

CLI flag: `--aws-access-key`, `--aws-secret-key`, `--aws-profile` and `--aws-region`

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

## `rule` blocks

CLI flag: `--ignore-rule`

You can make settings for each rule in the `rule` block. Currently, it can set only `enabled` option. If you set `enabled = false`, TFLint doesn't inspect configuration files by this rule.

```hcl
rule "aws_instance_previous_type" {
  enabled = false
}
```

You can also disable rules with the `--ignore-rule` option.

```
$ tflint --ignore-rule=aws_instance_invalid_type,aws_instance_previous_type
```

Also, annotation comments can disable rules on specific lines:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type
    instance_type = "t1.2xlarge"
}
```

The annotation works only for the same line or the line below it. You can also use `tflint-ignore: all` if you want to ignore all the rules.

See also [list of available rules](../rules).
