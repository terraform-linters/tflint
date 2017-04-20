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

// override by `template_override.tf`
resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t1.2xlarge"
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

resource "aws_cloudwatch_metric_alarm" "rds-writer-memory" {
  alarm_name                = "terraform-test-foobar5"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  unit                      = "percent" // aws_cloudwatch_metric_alarm_invalid_unit
  alarm_description         = "This metric monitor ec2 cpu utilization"
  insufficient_data_actions = []
}

module "ec2_instance" {
  source         = "github.com/wata727/example-module" // terraform_module_pinned_source
  instance_types = "${var.instance_type}"
  mysql_types    = "${var.mysql_type}"
  redis_previous = "${var.redis_previous_type}"
  redis_invalid  = "${var.redis_invalud_type}"
}
