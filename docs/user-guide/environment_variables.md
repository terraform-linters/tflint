# Environment Variables

Below is a list of environment variables available in TFLint.

- `TFLINT_LOG`
  - Print logs to stderr. See [Debugging](../../README.md#debugging).
- `TFLINT_CONFIG_FILE`
  - Configure the config file path. See [Configuring TFLint](./config.md).
- `TFLINT_PLUGIN_DIR`
  - Configure the plugin directory. See [Configuring Plugins](./plugins.md).
- `TFLINT_DISABLE_VERSION_CHECK`
  - Disable version update notifications when running `tflint --version`. Set to `1` to disable.
- `TFLINT_EXPERIMENTAL`
  - Enable experimental features. Note that experimental features are subject to change without notice. Currently only [Keyless Verification](./plugins.md#keyless-verification-experimental) are supported.
- `TF_VAR_name`
  - Set variables for compatibility with Terraform. See [Compatibility with Terraform](./compatibility.md).
- `TF_DATA_DIR`
  - Configure the `.terraform` directory for compatibility with Terraform. See [Compatibility with Terraform](./compatibility.md).
- `TF_WORKSPACE`
  - Set a workspace for compatibility with Terraform. See [Compatibility with Terraform](./compatibility.md).
