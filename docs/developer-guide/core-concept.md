# Core Concept

TFLint is just a thin wrapper of Terraform. Configuration loading and expression evaluation etc. depend on Terraform's internal API, and it only provides an interface to do them as a linter.

Rules are provided by plugins except some rules. Technically, the plugin is launched as another process, communicates via RPC, and receives inspection results from the plugin process.

There are important packages to understand its behavior:

- `tflint`
  - This package is the core of TFLint as a wrapper for Terraform. It allows accesses to `terraform/configs.Config` and `terraform/terraform.BuiltinEvalContext` and so on.
- `plugin`
  - This package provides the TFLint plugin system. Includes plugin discovery, a server implementation responding to requests from plugins.
- `cmd`
  - This package is the entrypoint of the app.
