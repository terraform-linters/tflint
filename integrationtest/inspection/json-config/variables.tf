# Standard module structure test file
# This file should satisfy the terraform_standard_module_structure rule

variable "proper_variable" {
  type        = string
  description = "A properly named variable"
  default     = "default_value"
}

variable "another_variable" {
  type = number
  description = "Another properly named variable"
  default = 42
}

# This variable name should trigger terraform_naming_convention rule
variable "ImproperVariable" {
  type        = string
  description = "A test variable with improper naming"
  default     = "test"
}