module "root" {
  source = "./module"

  invalid = "${terraform.env}"
}
