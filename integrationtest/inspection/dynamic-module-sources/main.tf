variable "prefix" {
  const = true
  default = "module"
}

variable "unknown" {
  const = true
}

locals {
  suffix = "suffix"
}

module "module" {
  source = "${path.module}/${var.prefix}-${local.suffix}" # should be "./module-suffix"

  nested_src = "${path.module}/nested"                    # should be "./nested"
}

module "unknown" {
  source = var.unknown # should be ignored
}

module "count_zero" {
  count = 0

  source = "${path.module}/${var.prefix}-${local.suffix}"

  nested_src = var.prefix.invalid # the error should be ignored because count = 0
}
