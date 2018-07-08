variable "string_var" {}
variable "integer_var" {}
variable "list_var" {}
variable "map_var" {}
variable "no_value_var" {}

resource "null_resource" "test" {
  // string
  literal      = "literal_val"
  string       = "${var.string_var}"
  new_string   = var.string_var
  list_element = "${var.list_var[0]}"
  map_element  = "${var.map_var["one"]}"
  conditional  = "${true ? "production" : "development"}"
  function     = "${md5("foo")}"
  workspace    = "${terraform.workspace}"
  inside       = "Hello ${var.string_var}"

  // integer
  integer = "${var.integer_var}"

  // list
  string_list = ["one", "two", "three"]
  number_list = [1, 2, 3]

  // map
  map = {
    one = 1
    two = 2
  }

  // error
  undefined = "${var.undefined_var}"
  no_value  = "${var.no_value_var}"
  env       = "${terraform.env}"

  // evaluable
  module      = "${module.text}"
  resource    = "${aws_subnet.app.id}"
  unsupported = "${var.text} ${lookup(var.roles, count.index)}"
}
