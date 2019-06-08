# TFLint Developer's Guide

The goal of this guide is to quickly understand how TFLint works and to help more developers contribute.

## Core Concept

TFLint is just a thin wrapper of Terraform. Configuration loading and expression evaluation etc. depend on Terraform's internal API, and it only provides an interface to do them as linter.

There are three important packages to understand its behavior:

- `tflint`
  - This package is the core of TFLint as a wrapper for Terraform. It allows accesses to `terraform/configs.Config` and `terraform/terraform.BuiltinEvalContext` and so on.
- `rules`
  - This package is a provider of all rules.
- `cmd`
  - This package is the entrypoint of the app.

## How does it work

These processes are described in [`cmd/cli.go`](https://github.com/wata727/tflint/blob/master/cmd/cli.go).

### 1. Loading configurations

All Terraform's configuration files are represented as `configs.Config`. [`tflint/tflint.Loader`](https://github.com/wata727/tflint/blob/master/tflint/loader.go) uses the `(*configs.Parser) LoadConfigDir` and `configs.BuildConfig` to access to `configs.Config` in the same way as Terraform.

Similarly, prepare `terraform.InputValues` using `(*configs.Parser) LoadValuesFile`.

### 2. Setting up a new Runner

A [`tflint/tflint.Runner`](https://github.com/wata727/tflint/blob/master/tflint/runner.go) is initialized for each `configs.Config`. These have their own evaluation context for that module, represented as `terraform.BuiltinEvalContext`.

it uses `(*terraform.BuiltinEvalContext) EvaluateExpr` to evaluate expressions. Unlike Terraform, it provides a mechanism to determine if an expression can be evaluated.

### 3. Inspecting configurations

It inspects `configs.Config` via `tflint.Runner`. All rules implement the `Check` method that takes `tflint.Runner` as an argument, and emits an issue if needed.

## Building

You need Go 1.12 or later to build.

```
$ make build
```

## Adding a new rule

You can use the rule generator to add new rules (Currently, this generator supports only AWS rules).

```
$ make rule
go run tools/rule_generator.go
Rule name? (e.g. aws_instance_invalid_type): aws_instance_example
Create: rules/awsrules/aws_instance_example.go
Create: rules/awsrules/aws_instance_example_test.go
```

A template of rules and tests is generated. In order to inspect configuration files, you need to understand [the Runner API](https://github.com/wata727/tflint/blob/master/tflint/runner.go).

Finally, don't forget to register the created rule with [the provider](https://github.com/wata727/tflint/blob/master/rules/provider.go). After that the rule you created is enabled in TFLint.


## Committing

Before commit, please install [pre-commit](https://pre-commit.com/) and install pre-commit hooks by running `pre-commit install` in root directory of your local checkout.
