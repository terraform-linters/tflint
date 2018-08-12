module "root" {
  source = "./module"

  override = "foo"
  no_default = "bar"
  unknown = "${data.aws.ami.id}"
}
