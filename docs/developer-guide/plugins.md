# Writing Plugins

If you want to add custom rules, you can write ruleset plugins.

## Overview

Plugins are independent binaries and use [go-plugin](https://github.com/hashicorp/go-plugin) to communicate with TFLint over gRPC. TFLint executes the binary when the plugin is enabled, and the plugin process must act as an gRPC server for TFLint.

If you want to create a new plugin, [The template repository](https://github.com/terraform-linters/tflint-ruleset-template) is available to satisfy these specification. You can create your own repository from "Use this template" and easily add rules based on some reference rules.

The template repository uses the [SDK](https://github.com/terraform-linters/tflint-plugin-sdk) that wraps the go-plugin for communication with TFLint. See also the [Architecture](https://github.com/terraform-linters/tflint-plugin-sdk#architecture) section for the architecture of the plugin system.

## 1. Creating a repository from the template

Visit [tflint-ruleset-template](https://github.com/terraform-linters/tflint-ruleset-template) and click the "Use this template" button. Repository name must be `tflint-ruleset-*`.

For public repositories, you can optionally add the [`tflint-ruleset`](https://github.com/topics/tflint-ruleset) topic to make your ruleset discoverable on the topic page.

## 2. Building and installing the plugin

The created repository can be installed locally with `make install`. Enable the plugin as follows and verify that the installed plugin works.

```hcl
plugin "template" {
    enabled = true
}
```

```console
$ make install
go build
mkdir -p ~/.tflint.d/plugins
mv ./tflint-ruleset-template ~/.tflint.d/plugins
$ tflint -v
TFLint version 0.28.1
+ ruleset.template (0.1.0)
```

## 3. Changing/Adding the rules

Rename the ruleset and add/edit rules. After making changes, you can check the behavior with `make install`. See also the [tflint-plugin-sdk API reference](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk) for communication with the host process.

## 4. Creating a GitHub Release

You can build and install your own ruleset locally as described above, but you can also install it automatically with `tflint --init`.

The requirements to support automatic installation are as follows:

- The built plugin binaries must be published on GitHub Release
- The release must be tagged with a name like `v1.1.1`
- The release must contain an asset with a name like `tflint-ruleset-{name}_{GOOS}_{GOARCH}.zip`
- The zip file must contain a binary named `tflint-ruleset-{name}` (`tflint-ruleset-{name}.exe` in Windows)
- The release must contain a checksum file for the zip file with the name `checksums.txt`
- The checksum file must contain a sha256 hash and filename

When signing a release, one of the following requirements must be met:

- [Artifact Attestation](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds)
  - Include a step in your GitHub Actions workflow that uses the [`attest-build-provenance` action](https://github.com/actions/attest-build-provenance) for `checksums.txt`.
- PGP signing key
  - The release must contain a signature file for the checksum file with the name `checksums.txt.sig`
  - The signature file must be binary OpenPGP format

Releases that meet these requirements can be easily created by following the GoReleaser config in the template repository.
