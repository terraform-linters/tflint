## What's Changed

### Breaking Changes
* plugin: Drop support for plugin SDK v0.14/v0.15 by @wata727 in https://github.com/terraform-linters/tflint/pull/2203
  * Plugins built with SDKs v0.14/v0.15 are no longer supported. If you get "SDK version is incompatible" error, you need to update the plugin to use SDK v0.16+.

### Enhancements
* Move recursive init output to debug when there are no changes by @pvickery-ParamountCommerce in https://github.com/terraform-linters/tflint/pull/2150
* Introduce plugin keyless verification by @wata727 in https://github.com/terraform-linters/tflint/pull/2159
  * For third-party plugins that are not PGP signed and have uploaded [artifact attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds), TFLint will now attempt to verify them via the Sigstore ecosystem.
* Add support for Terraform v1.10 by @wata727 in https://github.com/terraform-linters/tflint/pull/2178
* cmd: Simplify recursive init outputs by @wata727 in https://github.com/terraform-linters/tflint/pull/2204

### Chores
* build(deps): Bump goreleaser/goreleaser-action from 6.0.0 to 6.1.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2165
* build(deps): Bump actions/attest-build-provenance from 1.4.3 to 1.4.4 by @dependabot in https://github.com/terraform-linters/tflint/pull/2166
* build(deps): Bump golang.org/x/net from 0.30.0 to 0.31.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2164
* build(deps): Bump google.golang.org/grpc from 1.67.1 to 1.68.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2160
* build(deps): Bump golang.org/x/oauth2 from 0.23.0 to 0.24.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2163
* build(deps): Bump github.com/zclconf/go-cty from 1.15.0 to 1.15.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2174
* build(deps): Bump docker/build-push-action from 6.9.0 to 6.10.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2173
* build(deps): Bump mislav/bump-homebrew-formula-action from 3.1 to 3.2 by @dependabot in https://github.com/terraform-linters/tflint/pull/2171
* build(deps): Bump docker/metadata-action from 5.5.1 to 5.6.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2170
* build(deps): Bump github.com/hashicorp/hcl/v2 from 2.22.0 to 2.23.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2168
* build(deps): Bump github.com/theupdateframework/go-tuf/v2 from 2.0.0 to 2.0.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2177
* install: handle GitHub API JSON without newlines by @bendrucker in https://github.com/terraform-linters/tflint/pull/2176
* build(deps): Bump alpine from 3.20 to 3.21 by @dependabot in https://github.com/terraform-linters/tflint/pull/2180
* build(deps): Bump google.golang.org/grpc from 1.68.0 to 1.68.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2181
* build(deps): Bump actions/attest-build-provenance from 1.4.4 to 2.0.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/2185
* build(deps): Bump golang.org/x/crypto from 0.29.0 to 0.30.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2183
* build(deps): Bump golang.org/x/net from 0.31.0 to 0.32.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2184
* build(deps): Bump alpine to 3.21 and golangci-lint to 1.62.2 by @chenrui333 in https://github.com/terraform-linters/tflint/pull/2188
* build(deps): Bump golang.org/x/crypto from 0.30.0 to 0.31.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2189
* build(deps): Bump actions/attest-build-provenance from 2.0.1 to 2.1.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2190
* build(deps): Bump docker/setup-buildx-action from 3.7.1 to 3.8.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2191
* build(deps): Bump google.golang.org/grpc from 1.68.1 to 1.69.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2192
* build(deps): Bump google.golang.org/grpc from 1.69.0 to 1.69.2 by @dependabot in https://github.com/terraform-linters/tflint/pull/2196
* build(deps): Bump golang.org/x/net from 0.32.0 to 0.33.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2195
* chore: fix typos by @chenrui333 in https://github.com/terraform-linters/tflint/pull/2198
* build(deps): Bump github.com/zclconf/go-cty from 1.15.1 to 1.16.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2199
* build(deps): Bump golang.org/x/crypto from 0.31.0 to 0.32.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2200
* build(deps): Bump golang.org/x/oauth2 from 0.24.0 to 0.25.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/2202

## New Contributors
* @pvickery-ParamountCommerce made their first contribution in https://github.com/terraform-linters/tflint/pull/2150

**Full Changelog**: https://github.com/terraform-linters/tflint/compare/v0.54.0...v0.55.0
