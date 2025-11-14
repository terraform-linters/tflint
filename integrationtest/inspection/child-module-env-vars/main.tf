terraform {
  required_version = ">= 1.7.0"
}

variable "environment" {
  type = string
}

module "child" {
  source = "./child-module"

  environment = [var.environment]
}

output "env_output" {
  value = var.environment
}
