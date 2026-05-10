variable "nested_src" {
  const = true
}

module "nested" {
  source = var.nested_src

  input = var.nested_src
}
