## What's Changed

This release includes many new features including parallelization of recursion inspection and support for Terraform v1.8.

Also, please be aware that there are important changes regarding licensing. TFLint has updated the embedded Terraform package to the latest version for Terraform v1.6+ support. As a result, we will be affected by [Terraform's license change to BUSL announced by Hashicorp in August 2023](https://www.hashicorp.com/blog/hashicorp-adopts-business-source-license).

Most of the code in TFLint is still licensed under MPL 2.0, but some files under the Terraform package are now licensed under BUSL 1.1. This means that release binaries are bound by both licenses and may be subject to Hashicorp's BUSL restrictions. If you have concerns about this change, we recommend reviewing the licensing implications before updating. Please note that we cannot provide legal advice.

Please refer to the discussion in https://github.com/terraform-linters/tflint/discussions/1826 and https://github.com/terraform-linters/tflint/issues/1878 for details.

### Enhancements

* config: Add TFLint `required_version` settings by @wata727 in https://github.com/terraform-linters/tflint/pull/2027
  * The `required_version` attribute can now be set in `.tflint.hcl`. This is useful for enforcing the version of TFLint that is actually used.
* plugin: Add support for host-specific GitHub tokens by @wata727 in https://github.com/terraform-linters/tflint/pull/2025
  * Environment variables like `GITHUB_TOKEN_example_com` have been introduced for GitHub Enterprise Server support.
* Recursive inspection in parallel by @wata727 in https://github.com/terraform-linters/tflint/pull/2021
  * The `--recursive` inspection now runs in parallel according to the number of CPU cores by default. The number of parallels can be changed with `--max-workers`.
* terraform: Add support for Terraform v1.6/v1.7/v1.8 by @wata727 in https://github.com/terraform-linters/tflint/pull/2030
  * New Terraform features are now supported, including [provider-defined functions](https://www.hashicorp.com/blog/terraform-1-8-adds-provider-functions-for-aws-google-cloud-and-kubernetes). Please note that support for provider-defined functions requires the latest HCL parser, so you may need to update your plugin versions.
  * Updated embedded Terraform packages to support Terraform v1.6+. As a result, TFLint now includes code for Hashicorp's BUSL 1.1.

### Changes

* Add warnings to --module/--no-module and `module` attribute by @wata727 in https://github.com/terraform-linters/tflint/pull/1951
  * If you see a warning, use `--call-module-type` instead. The `--module` is equivalent to `--call-module-type=all` and the `--no-module` is equivalent to `--call-module-type=none`. This also applies to `.tflint.hcl`.

### Chores

* build: use go1.22 by @chenrui333 in https://github.com/terraform-linters/tflint/pull/1977
* workflows: remove `cache: true` for setup-go (default) by @chenrui333 in https://github.com/terraform-linters/tflint/pull/1979
* install: enable `pipefail` catch `curl` errors by @Ry4an in https://github.com/terraform-linters/tflint/pull/1978
* build(deps): Bump golang.org/x/oauth2 from 0.16.0 to 0.17.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1981
* build(deps): Bump golangci/golangci-lint-action from 3.7.0 to 4.0.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1980
* build(deps): Bump google.golang.org/grpc from 1.61.0 to 1.61.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/1987
* sarif: add schema to repo by @bendrucker in https://github.com/terraform-linters/tflint/pull/2000
* build(deps): Bump google.golang.org/grpc from 1.61.1 to 1.62.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1992
* build(deps): Bump github.com/hashicorp/hcl/v2 from 2.19.1 to 2.20.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1999
* build(deps): Bump github.com/zclconf/go-cty from 1.14.2 to 1.14.3 by @dependabot in https://github.com/terraform-linters/tflint/pull/1998
* build(deps): Bump golang.org/x/crypto from 0.19.0 to 0.21.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2001
* build(deps): Bump golang.org/x/oauth2 from 0.17.0 to 0.18.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2002
* build(deps): Bump google.golang.org/grpc from 1.62.0 to 1.62.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2003
* build(deps): Bump github.com/zclconf/go-cty from 1.14.3 to 1.14.4 by @dependabot in https://github.com/terraform-linters/tflint/pull/2009
* build(deps): Bump github.com/hashicorp/hcl/v2 from 2.20.0 to 2.20.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2012
* build(deps): Bump google.golang.org/grpc from 1.62.1 to 1.63.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2014
* build(deps): Bump golang.org/x/crypto from 0.21.0 to 0.22.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2016
* build(deps): Bump golang.org/x/oauth2 from 0.18.0 to 0.19.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2015
* build(deps): Bump sigstore/cosign-installer from 3.4.0 to 3.5.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2022
* build(deps): Bump google.golang.org/grpc from 1.63.0 to 1.63.2 by @dependabot in https://github.com/terraform-linters/tflint/pull/2023
* build(deps): Bump golang.org/x/net from 0.22.0 to 0.23.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2024
* build(deps): Bump github.com/hashicorp/go-getter from 1.7.2 to 1.7.4 by @dependabot in https://github.com/terraform-linters/tflint/pull/2026
* build(deps): Bump golangci/golangci-lint-action from 4.0.0 to 5.1.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2029
* Pin Go patch version in go.mod by @wata727 in https://github.com/terraform-linters/tflint/pull/2031
* build(deps): Bump github.com/terraform-linters/tflint-plugin-sdk from 0.18.0 to 0.20.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2032
* build(deps): Bump github.com/terraform-linters/tflint-ruleset-terraform from 0.5.0 to 0.7.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2033

## New Contributors
* @Ry4an made their first contribution in https://github.com/terraform-linters/tflint/pull/1978

**Full Changelog**: https://github.com/terraform-linters/tflint/compare/v0.50.3...v0.51.0
