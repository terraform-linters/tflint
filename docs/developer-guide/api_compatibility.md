# API Compatibility

This is an internal documentation summarizing the currently supported SDK and TFLint versions and any compatibility caveats.

Protocol version: 11  
SDK version: v0.14.0+
TFLint version: v0.42.0+

- Client-side value handling is introduced in SDK v0.16.0 and TFLint v0.46.0. TFLint v0.45.0 returns an error instead of a value.
  - https://github.com/terraform-linters/tflint/pull/1700
  - https://github.com/terraform-linters/tflint/pull/1722
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/235
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/239
- Ephemeral value is introduced in SDK v0.22.0 and TFLint v0.55.0. TFLint returns ErrSensitive instead of ephemeral values to plugins built with SDK v0.21.0.
  - https://github.com/terraform-linters/tflint/pull/2178
  - https://github.com/terraform-linters/tflint-plugin-sdk/pull/358
