variable "foo" {}
variable "bar" {}
variable "baz" {}

module "module2" {
  source = "./module"

  red   = "${var.foo}-${var.bar}"
  blue  = "blue"
  green = "${var.foo}-${var.baz}-${path.module}"
}
