terraform {
  required_version = ">= 1.7.0"
}

variable "environment" {
  type = list(string)
}

output "child_env" {
  value = var.environment
}
