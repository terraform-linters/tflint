variable "invalid" {
  default = "foo"
}

resource "null_resource" "null" {
  foo = "${var.invalid}"
}
