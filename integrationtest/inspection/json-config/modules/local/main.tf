# Local module for testing
variable "input" {
  type = string
  default = "test"
}

output "output" {
  value = var.input
}