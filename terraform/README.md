# Forked Terraform packages

This directory contains a subset of code from Terraform's internal packages. However, the implementation is not exactly the same, it is just a fork, and simplifications and changes have been made according to our project.

Previously, TFLint uses Terraform as a library, but due to the [package internalization](https://github.com/hashicorp/terraform/issues/26418), it is no longer possible to import from external packages such as TFLint. This is why TFLint has its own fork.

This package provides functionality for static analysis of Terraform Language.
