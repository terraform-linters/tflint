# Advanced Inspection

## Deep Checking

When deep checking is enabled, TFLint invokes the provider's API to do a more detailed inspection. For example, find a non-existent IAM profile name etc. You can enable it with the `--deep` option.

```console
$ tflint --deep
2 issue(s) found:

Error: instance_type is not a valid value (aws_instance_invalid_type)

  on template.tf line 3:
   3:   instance_type        = "t1.2xlarge"

Error: "invalid_profile" is invalid IAM profile name. (aws_instance_invalid_iam_profile)

  on template.tf line 4:
   4:   iam_instance_profile = "invalid_profile"

```

In order to enable deep checking, [credentials](credentials.md) are needed.

## Module Inspection

TFLint can also inspect [modules](https://www.terraform.io/docs/configuration/modules.html). In this case, it checks based on the input variables passed to the calling module.

```hcl
module "aws_instance" {
  source        = "./module"

  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge"
}
```

```console
$ tflint --module
1 issue(s) found:

Error: instance_type is not a valid value (aws_instance_invalid_type)

  on template.tf line 5:
   5:   instance_type = "t1.2xlarge"

Callers:
   template.tf:5,19-31
   module/instance.tf:5,19-36

```

Module inspection is disabled by default. Inspection is enabled by running with the `--module` option. Note that you need to run `terraform init` first because of TFLint loads modules in the same way as Terraform. 

You can use the `--ignore-module` option if you want to skip inspection for a particular module. Note that you need to pass module sources rather than module ids for backward compatibility.

```
$ tflint --ignore-module=./module
```

## Only Mode

TFLint allows you to specifically enable *only* certain rules, and disable all other rules including the default ruleset. This can be useful for splitting up linting workflows, by separating which rules to inspect at which stage.

For example, you might want a centralized tag-keys linter, checking that all taggable AWS resources contain a set of tags across multiple repositories. You might want to separate that from your other TFLint workflows, because each repo may also have a varying set of rule configurations it wants to apply.

To use Only Mode, you can pass rules on the CLI like so:
```
$ tflint --only aws_instance_invalid_type --only aws_instance_invalid_ami
```
**Note:** usage of `--only` will ignore any other rules defined via command line via `--enable-rule` or `--disable-rule`.

You can also set `disabled_by_default = true` in the config file. Using this method, any rules enabled in your config file will implicitly become `--only` rules, and all other defaults will be ignored.

```hcl
config {
  disabled_by_default = true
  # other options here...
}

rule "aws_instance_previous_type" {
  enabled = true
}
```