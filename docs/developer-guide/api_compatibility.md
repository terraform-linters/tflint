# API Compatibility

This is an internal documentation summarizing the currently supported SDK and TFLint versions and any compatibility caveats.

Protocol version: 11  
SDK version: v0.12.0+  
TFLint version: v0.40.0+  

- `Only` option is only supported by SDK v0.13.0+.
  - https://github.com/terraform-linters/tflint/pull/1516
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/198
- Schema mode is only supported by SDK v0.14.0+ and TFLint v0.42.0+. v0.41 ignores this mode.
  - https://github.com/terraform-linters/tflint/pull/1530
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/201
- `VersionConstraint` is only supported by SDK v0.14.0+ and TFLint v0.42.0+. v0.41 ignores this constraint.
  - https://github.com/terraform-linters/tflint/pull/1535
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/202
- `SDKVersion` is only supported by SDK v0.14.0+. v0.13 does not return a version.
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/203
- `each.*` and `count.*` are only supported by SDK v0.14.0+ and TFLint v0.42.0+.
  - https://github.com/terraform-linters/tflint/pull/1537
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/205
- Expand mode is only supported by SDK v0.14.0+ and TFLint v0.42.0+.
  - https://github.com/terraform-linters/tflint/pull/1537
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/208
