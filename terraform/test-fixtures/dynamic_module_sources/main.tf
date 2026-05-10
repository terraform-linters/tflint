variable "module_src" {
  const   = true
  default = "./modules"
}

variable "unknown" {
  const = true
}

variable "sensitive" {
  const     = true
  default   = "./modules"
  sensitive = true
}

variable "null" {
  const   = true
  default = null
}

module "module" {
  source = var.module_src

  nested_src = "./nested"
}

module "unknown" {
  source = var.unknown
}

module "count_zero" {
  count = 0

  source = var.module_src

  nested_src = var.module_src.invalid
}

module "sensitive" {
  source = var.sensitive
}

module "null" {
  source = var.null
}
