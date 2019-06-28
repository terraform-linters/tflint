# Limitations

This page describes some of the limitations of TFLint inspection.

## Supported Providers

Currently only supported Terraform itself and AWS provider.

## Supported Versions

Some inspections implicitly assume the behavior of a specific version of provider plugins or Terraform. This always assumes the latest version and is as follows:

- Terraform v0.12.3
- AWS Provider v2.16.0

Of course, TFLint may work correctly if you run it on other versions. But, false positives/negatives can occur based on this assumption.

## Supported Named Values

[Named values](https://www.terraform.io/docs/configuration/expressions.html#references-to-named-values) are supported only for [input variables](https://www.terraform.io/docs/configuration/variables.html) and [workspaces](https://www.terraform.io/docs/state/workspaces.html). Expressions that contain anything else are excluded from the  inspection. [Built-in Functions](https://www.terraform.io/docs/configuration/functions.html) are fully supported.
