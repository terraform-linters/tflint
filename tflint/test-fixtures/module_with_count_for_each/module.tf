variable "config" {
  type = object({ instance_type = string })
  default = null
}

module "count_is_zero" {
  source = "./module"
  count = var.config != null ? 1 : 0

  instance_type = var.config.instance_type
}

module "count_is_one" {
  source = "./module"
  count = var.config != null ? 0 : 1

  instance_type = "t2.micro"
}

variable "instance_types" {
  type = list(string)
  default = []
}

module "for_each_is_empty" {
  source = "./module"
  for_each = var.instance_types

  instance_type = each.value
}

variable "instance_types_with_default" {
  type = list(string)
  default = ["t2.micro"]
}

module "for_each_is_not_empty" {
  source = "./module"
  for_each = var.instance_types_with_default

  instance_type = each.value
}
