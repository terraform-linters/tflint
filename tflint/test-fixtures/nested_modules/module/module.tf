variable "override" {
  default = "baz"
}
variable "no_default" {}
variable "unknown" {}

module "test" {
  source = "./module1"

  override = "${var.override}"
  no_default = "${var.no_default}"
  unknown = "${var.unknown}"
}
