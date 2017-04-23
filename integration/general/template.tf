provider "aws" {
  region = "us-east-1"
}

variable "instance_type" {
  type = "map"
}

variable "mysql_type" {
  type = "list"
}

variable "redis_invalud_type" {
  type = "string"
}

variable "redis_previous_type" {
  type    = "string"
  default = "cache.t2.micro"
}

resource "aws_route" "not_specified" {
  route_table_id         = "rtb-1234abcd" // aws_route_not_specified_target
  destination_cidr_block = "10.0.1.0/22"
}

resource "aws_route" "multiple_specified" {
  route_table_id         = "rtb-1234abcd"  // aws_route_specified_multiple_targets
  destination_cidr_block = "10.0.1.0/22"
  gateway_id             = "igw-1234abcd"
  egress_only_gateway_id = "eigw-1234abcd"
}

module "ec2_instance" {
  source         = "github.com/wata727/example-module" // terraform_module_pinned_source
  instance_types = "${var.instance_type}"
  mysql_types    = "${var.mysql_type}"
  redis_previous = "${var.redis_previous_type}"
  redis_invalid  = "${var.redis_invalud_type}"
}
