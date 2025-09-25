# This module should be ignored by TFLint

variable "ignored_input" {
  type = string
}

output "ignored_output" {
  value = var.ignored_input
}