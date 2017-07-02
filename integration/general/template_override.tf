resource "aws_instance" "web" {
  ami                  = "ami-12345678"
  instance_type        = "t2.2xlarge"
  iam_instance_profile = "web-server"
}

module "ec2_instance" {
  source         = "github.com/wata727/example-module" // terraform_module_pinned_source
  instance_types = "${var.instance_type}"
  mysql_types    = "${var.mysql_type}"
  redis_previous = "${var.redis_previous_type}"
  redis_invalid  = "${var.redis_invalud_type}"
}
