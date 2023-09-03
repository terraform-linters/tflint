## What's Changed

### Breaking Changes

* Bump tflint-plugin-sdk to v0.18.0 by @wata727 in https://github.com/terraform-linters/tflint/pull/1813
  * This change causes the deprecated `IncludeNotCreated` option to be ignored. Most plugin users will not be affected.

### BugFixes

* langserver: Trap os.Interrupt and syscall.SIGTERM by @wata727 in https://github.com/terraform-linters/tflint/pull/1809
* Bump github.com/hashicorp/hcl to v2.18.0 by @wata727 in https://github.com/terraform-linters/tflint/pull/1833
* tflint: Allow commas with spaces in annotations by @wata727 in https://github.com/terraform-linters/tflint/pull/1834

### Chores

* build(deps): Bump alpine from 3.18.0 to 3.18.2 by @dependabot in https://github.com/terraform-linters/tflint/pull/1784
* build(deps): Bump google.golang.org/grpc from 1.55.0 to 1.56.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1785
* build(deps): Bump golang.org/x/oauth2 from 0.8.0 to 0.9.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1786
* build(deps): Bump sigstore/cosign-installer from 3.0.5 to 3.1.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1792
* build(deps): Bump google.golang.org/grpc from 1.56.0 to 1.56.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/1793
* build(deps): Bump sigstore/cosign-installer from 3.1.0 to 3.1.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/1798
* Remove hard-coded versions from integration tests by @wata727 in https://github.com/terraform-linters/tflint/pull/1799
* build(deps): Bump golang.org/x/text from 0.10.0 to 0.11.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1806
* build(deps): Bump golang.org/x/crypto from 0.10.0 to 0.11.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1804
* build(deps): Bump golang.org/x/oauth2 from 0.9.0 to 0.10.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1803
* build(deps): Bump google.golang.org/grpc from 1.56.1 to 1.56.2 by @dependabot in https://github.com/terraform-linters/tflint/pull/1805
* Remove obsoleted PGP public key by @wata727 in https://github.com/terraform-linters/tflint/pull/1800
* Add make release for release automation by @wata727 in https://github.com/terraform-linters/tflint/pull/1802
* build(deps): Bump google.golang.org/grpc from 1.56.2 to 1.57.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1815
* build(deps): Bump golang.org/x/crypto from 0.11.0 to 0.12.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1820
* build(deps): Bump golang.org/x/text from 0.11.0 to 0.12.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1821
* build(deps): Bump golang.org/x/oauth2 from 0.10.0 to 0.11.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1822
* deps: upgrade to use go1.21 by @chenrui333 in https://github.com/terraform-linters/tflint/pull/1823
* build(deps): Bump github.com/google/uuid from 1.3.0 to 1.3.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/1829
* build(deps): Bump golangci/golangci-lint-action from 3.6.0 to 3.7.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1830


**Full Changelog**: https://github.com/terraform-linters/tflint/compare/v0.47.0...v0.48.0
