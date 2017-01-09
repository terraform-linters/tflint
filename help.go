package main

const Help string = `TFLint is a linter of Terraform.

Usage: tflint [<options>] <args>

Available options:
    -h, --help                              show usage of TFLint. This page.
    -v, --version                           print version information.
    -f, --format <format>                   choose output format from "default" or "json"
    -c, --config <file>                     specify config file. default is ".tflint.hcl"
    --ignore-module <source1,source2...>    ignore module by specified source.
    --ignore-rule <rule1,rule2...>          ignore rules.
    --deep                                  enable deep check mode.
    --aws-access-key                        set AWS access key used in deep check mode.
    --aws-secret-key                        set AWS secret key used in deep check mode.
    --aws-region                            set AWS region used in deep check mode.
    -d, --debug                             enable debug mode.
    --error-with-issues                     return error code when issue exists.

Support aruguments:
    TFLint scans all configuration file of Terraform in current directory by default.
    If you specified single file path, it scans only this.
`
