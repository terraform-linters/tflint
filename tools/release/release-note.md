## What's Changed

### Enhancements
* cmd: include worker dir for no-range recursive issues by @Zakharden in https://github.com/terraform-linters/tflint/pull/2524
* config: add ignorable rule setting by @Zakharden in https://github.com/terraform-linters/tflint/pull/2525
* formatter(junit): include range in testcase name by @Zakharden in https://github.com/terraform-linters/tflint/pull/2538

### Bug Fixes
* plugin: fetch attestation bundles from bundle_url by @Kunalbehbud in https://github.com/terraform-linters/tflint/pull/2593
  * Merged as https://github.com/terraform-linters/tflint/pull/2600

### Chores
* Refactor formatter dispatch into format adapters by @bendrucker in https://github.com/terraform-linters/tflint/pull/2556
* Inline per-format print methods into their adapters by @bendrucker in https://github.com/terraform-linters/tflint/pull/2557
* build(deps): Bump docker/setup-buildx-action from 4.0.0 to 4.1.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2559
* build(deps): Bump github.com/mattn/go-colorable from 0.1.14 to 0.1.15 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2560
* build(deps): Bump golang from 1.26.3-alpine3.23 to 1.26.4-alpine3.23 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2558
* release: Set branch name for Homebrew tap by @wata727 in https://github.com/terraform-linters/tflint/pull/2561
* Extract the shared runner-setup pipeline by @bendrucker in https://github.com/terraform-linters/tflint/pull/2555
* formatter: guard diagnostics with no source range by @bendrucker in https://github.com/terraform-linters/tflint/pull/2562
* docs: add Go install instructions to README by @RoseSecurity in https://github.com/terraform-linters/tflint/pull/2563
* build(deps): Bump actions/checkout from 6.0.2 to 6.0.3 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2565
* build(deps): Bump alpine from 3.23.4 to 3.24.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2566
* build(deps): Bump the go-x group with 2 updates by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2567
* build(deps): Bump github.com/sigstore/sigstore-go from 1.1.4 to 1.2.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2568
* fix(install): close download and zip readers by @RoseSecurity in https://github.com/terraform-linters/tflint/pull/2564
* build(deps): Bump alpine from 3.24.0 to 3.24.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2570
* build(deps): Bump golang.org/x/net from 0.55.0 to 0.56.0 in the go-x group by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2571
* build(deps): Bump github.com/sigstore/sigstore-go from 1.2.0 to 1.2.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2572
* build(deps): Bump actions/checkout from 6.0.3 to 7.0.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2579
* build(deps): Bump actions/setup-go from 6.4.0 to 6.5.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2580
* build(deps): Bump golangci/golangci-lint-action from 9.2.1 to 9.3.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2583
* build(deps): Bump goreleaser/goreleaser-action from 7.2.2 to 7.2.3 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2582
* build(deps): Bump actions/attest from 4.1.0 to 4.1.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2581
* build(deps): Bump docker/build-push-action from 7.2.0 to 7.3.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2584
* build(deps): Bump golang.org/x/text from 0.38.0 to 0.39.0 in the go-x group by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2585
* build(deps): Bump github.com/sigstore/sigstore-go from 1.2.1 to 1.2.2 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2588
* build(deps): Bump google.golang.org/grpc from 1.81.1 to 1.82.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint/pull/2587

## New Contributors
* @Zakharden made their first contribution in https://github.com/terraform-linters/tflint/pull/2524
* @RoseSecurity made their first contribution in https://github.com/terraform-linters/tflint/pull/2563
* @Kunalbehbud made their first contribution in https://github.com/terraform-linters/tflint/pull/2593

**Full Changelog**: https://github.com/terraform-linters/tflint/compare/v0.63.1...v0.64.0
