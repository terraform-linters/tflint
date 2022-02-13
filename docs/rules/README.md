# Rules

Rules are usually provided by ruleset plugins, but the rules for the Terraform Language are built into the TFLint binary. Terraform language rules implement recommendations from the [Terraform language documentation](https://www.terraform.io/language). If you want to enforce additional usage and style conventions in your configuration, you can author your own ruleset plugin.

Below is a list of available rules.

|Rule|Description|Enabled|
| --- | --- | --- |
|[terraform_comment_syntax](terraform_comment_syntax.md)|Disallow `//` comments in favor of `#`||
|[terraform_deprecated_index](terraform_deprecated_index.md)|Disallow legacy dot index syntax||
|[terraform_deprecated_interpolation](terraform_deprecated_interpolation.md)|Disallow deprecated (0.11-style) interpolation|✔|
|[terraform_documented_outputs](terraform_documented_outputs.md)|Disallow `output` declarations without description||
|[terraform_documented_variables](terraform_documented_variables.md)|Disallow `variable` declarations without description||
|[terraform_module_pinned_source](terraform_module_pinned_source.md)|Disallow specifying a git or mercurial repository as a module source without pinning to a version|✔|
|[terraform_module_version](terraform_module_version.md)|Checks that Terraform modules sourced from a registry specify a version|✔|
|[terraform_naming_convention](terraform_naming_convention.md)|Enforces naming conventions for resources, data sources, etc||
|[terraform_required_providers](terraform_required_providers.md)|Require that all providers have version constraints through required_providers||
|[terraform_required_version](terraform_required_version.md)|Disallow `terraform` declarations without require_version||
|[terraform_standard_module_structure](terraform_standard_module_structure.md)|Ensure that a module complies with the Terraform Standard Module Structure||
|[terraform_typed_variables](terraform_typed_variables.md)|Disallow `variable` declarations without type||
|[terraform_unused_declarations](terraform_unused_declarations.md)|Disallow variables, data sources, and locals that are declared but never used||
|[terraform_unused_required_providers](terraform_unused_required_providers.md)|Check that all `required_providers` are used in the module||
|[terraform_workspace_remote](terraform_workspace_remote.md)|`terraform.workspace` should not be used with a "remote" backend with remote execution|✔|
