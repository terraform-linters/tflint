# terraform_naming_convention

Enforces naming conventions for the following blocks:

* Resources
* Input variables
* Output values
* Local values
* Modules
* Data sources

## Configuration

Name | Default | Value
--- | --- | ---
enabled | `false` | Boolean
format | `snake_case` | `snake_case`, `mixed_snake_case`, `""`
custom | `""` | String representation of a golang regular expression that the block name must match
data | | Block settings to override naming convention for data sources
locals | | Block settings to override naming convention for local values
module | | Block settings to override naming convention for modules
output | | Block settings to override naming convention for output values
resource | | Block settings to override naming convention for resources
variable | | Block settings to override naming convention for input variables


#### `format`

The `format` option defines the allowed formats for the block label. 
This option accepts one of the following values:

* `snake_case` - standard snake_case format - all characters must be lower-case, and underscores are allowed.
* `mixed_snake_case` - modified snake_case format - characters may be upper or lower case, and underscores are allowed.
* `""` - empty string - signifies "this selector shall not have its format checked".
This can be useful if you want to enforce no particular format for a block.

#### `custom`

The `custom` option defines a custom regex that the identifier must match. This option allows you to have a 
bit more finer-grained control over identifiers, letting you force certain patterns and substrings.

## Examples

### Default - enforce snake_case for all blocks

#### Rule configuration

```
rule "terraform_naming_convention" {
  enabled = true
}
```

#### Sample terraform source file

```hcl
data "aws_eip" "camelCase" {
}

data "aws_eip" "valid_name" {
}
```

```
$ tflint
1 issue(s) found:

Notice: data name `camelCase` must match the following format: snake_case (terraform_naming_convention)

  on template.tf line 1:
   1: data "aws_eip" "camelCase" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```


### Custom naming expression for all blocks

#### Rule configuration

```
rule "terraform_naming_convention" {
  enabled = true

  custom = "^[a-zA-Z]+([_-][a-zA-Z]+)*$"
}
```

#### Sample terraform source file

```hcl
resource "aws_eip" "Invalid_Name_With_Number123" {
}

resource "aws_eip" "Name-With_Dash" {
}
```

```
$ tflint
1 issue(s) found:

Notice: resource name `Invalid_Name_With_Number123` must match the following RegExp: ^[a-zA-Z]+([_-][a-zA-Z]+)*$ (terraform_naming_convention)

  on template.tf line 1:
   1: resource "aws_eip" "Invalid_Name_With_Number123" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```


### Override default setting for specific block type

#### Rule configuration

```
rule "terraform_naming_convention" {
  enabled = true

  module {
    custom = "^[a-zA-Z]+(_[a-zA-Z]+)*$"
  }
}
```

#### Sample terraform source file

```hcl
// data name enforced with default snake_case
data "aws_eip" "eip_1a" {
}

module "valid_module" {
  source = ""
}

module "invalid_module_with_number_1a" {
  source = ""
}
```

```
$ tflint
1 issue(s) found:

Notice: module name `invalid_module_with_number_1a` must match the following RegExp: ^[a-zA-Z]+(_[a-zA-Z]+)*$ (terraform_naming_convention)

  on template.tf line 9:
   9: module "invalid_module_with_number_1a" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```

### Disable for specific block type

#### Rule configuration

```
rule "terraform_naming_convention" {
  enabled = true

  module {
    format = ""
  }
}
```

#### Sample terraform source file

```hcl
// data name enforced with default snake_case
data "aws_eip" "eip_1a" {
}

// module names will not be enforced
module "Valid_Name-Not-Enforced" {
  source = ""
}
```


### Disable for all blocks but enforce a specific block type

#### Rule configuration

```
rule "terraform_naming_convention" {
  enabled = true
  format  = ""

  local {
    format = "snake_case"
  }
}
```

#### Sample terraform source file

```hcl
// Data block name not enforced
data "aws_eip" "EIP_1a" {
}

// Resource block name not enforced
resource "aws_eip" "EIP_1b" {
}

// local variable names enforced
locals {
  valid_name   = "valid"
  invalid-name = "dashes are not allowed with snake_case"
}
```

```
$ tflint
1 issue(s) found:

Notice: local value name `invalid-name` must match the following format: snake_case (terraform_naming_convention)

  on template.tf line 12:
  12: invalid-name = "dashes are not allowed with snake_case"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```

## Why

Naming conventions are optional, so it is not necessary to follow this. 
But this rule is useful if you want to force the following naming conventions in line with the [Terraform Plugin Naming Best Practices](https://www.terraform.io/docs/extend/best-practices/naming.html).

## How To Fix

Update the block label according to the format or custom regular expression.
