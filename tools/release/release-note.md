## What's Changed

### Breaking Changes
* Call local modules by default by @wata727 in https://github.com/terraform-linters/tflint/pull/1918
  * Module inspection is now enabled by default for modules whose source is a relative path. Note that "module inspection" will be called "calling modules" after this change. See also https://github.com/terraform-linters/tflint/issues/1066
  * CLI flag `--module` has been changed to `--call-module-type`. For backward compatibility, `--module` will continue to work, but it will be removed in a future version, so we recommend migrating early. The same applies to the `module` attribute of the configuration file.
    * `--module` flag is replaced by `--call-module-type=all` and `--no-module` (previous default) is replaced by `--call-module-type=none`
  * For modules with many local module calls, this change may result in performance degradation. If this is not acceptable, you can keep the previous default by specifying `--call-module-type=none`.
* Make assignments to undeclared variables an error by @wata727 in https://github.com/terraform-linters/tflint/pull/1941
  * In line with Terraform behavior, assignments using the `--var` flag etc. to undeclared variables now result in an error. To avoid this, remove unnecessary variable assignments.

### Enhancements
* Print the working directory on error in recursive inspection by @wata727 in https://github.com/terraform-linters/tflint/pull/1933
* Enable per-runner parallelism by @wata727 in https://github.com/terraform-linters/tflint/pull/1944

### BugFixes
* Exit with an error if the explicitly passed `.tflint.hcl` does not exist by @wata727 in https://github.com/terraform-linters/tflint/pull/1940

### Chores
* build(deps): Bump golang.org/x/oauth2 from 0.13.0 to 0.14.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1913
* build(deps): Bump sigstore/cosign-installer from 3.1.2 to 3.2.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1915
* build(deps): Bump github.com/hashicorp/go-plugin from 1.5.2 to 1.6.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1917
* docs: Remove mention of directory arguments by @wata727 in https://github.com/terraform-linters/tflint/pull/1921
* build(deps): Bump golang.org/x/crypto from 0.15.0 to 0.16.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1923
* build(deps): Bump golang.org/x/oauth2 from 0.14.0 to 0.15.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1931
* build(deps): Bump github.com/spf13/afero from 1.10.0 to 1.11.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1932
* build(deps): Bump actions/setup-go from 4 to 5 by @dependabot in https://github.com/terraform-linters/tflint/pull/1936
* build(deps): Bump sigstore/cosign-installer from 3.2.0 to 3.3.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1937
* build(deps): Bump alpine from 3.18 to 3.19 by @dependabot in https://github.com/terraform-linters/tflint/pull/1938
* Stop using backticks for emphasis by @wata727 in https://github.com/terraform-linters/tflint/pull/1934
* Avoid escaping newlines by @wata727 in https://github.com/terraform-linters/tflint/pull/1942
* build(deps): Bump golang.org/x/crypto from 0.16.0 to 0.17.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1945
* build(deps): Bump github.com/google/uuid from 1.4.0 to 1.5.0 by @dependabot in https://github.com/terraform-linters/tflint/pull/1947
* build(deps): Bump google.golang.org/grpc from 1.59.0 to 1.60.1 by @dependabot in https://github.com/terraform-linters/tflint/pull/1948


**Full Changelog**: https://github.com/terraform-linters/tflint/compare/v0.49.0...v0.50.0
