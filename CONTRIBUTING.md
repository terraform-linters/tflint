# Contributing to TFLint

Thank you for your interest in contributing to TFLint! Before submitting a pull request, please read the licensing notes below.

See the [Developer Guide](docs/developer-guide) for how to build and test TFLint.

## Licensing

TFLint is licensed under two licenses:

- Most of this project is licensed under the [Mozilla Public License 2.0](LICENSE) (MPL 2.0).
- Some files in the `terraform` package are licensed under the [Business Source License 1.1](LICENSE-BUSL) (BUSL 1.1). Please note that BUSL 1.1 is **not** an open source license.

The license of each file is indicated by the `SPDX-License-Identifier` header in the file. Files without the header are licensed under MPL 2.0.

As stated in the [GitHub Terms of Service](https://docs.github.com/en/site-policy/github-terms/github-terms-of-service#6-contributions-under-repository-license), your contribution is licensed under the same license as the file you change or add.

### Terraform fork

The `terraform` package is a fork of [Terraform](https://github.com/hashicorp/terraform)'s internal packages. Upstream Terraform has been relicensed from MPL 2.0 to BUSL 1.1, and some files in the fork that are currently licensed under MPL 2.0 may need to be replaced with their BUSL 1.1 upstream versions in the future.

To keep that possible, if you submit changes to files under the `terraform` directory, you agree to the following terms in addition to the above:

> You grant the TFLint maintainers the right to relicense your contribution to files under the `terraform` directory under the Business Source License 1.1, if the corresponding upstream Terraform code is licensed under the Business Source License 1.1.

To indicate your agreement, please check the box in the pull request description. If you do not agree, your pull request may not be accepted.
