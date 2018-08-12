module "root" {
  source = "./module"

  invalid {
    aws = "1.1.3"
  }
}
