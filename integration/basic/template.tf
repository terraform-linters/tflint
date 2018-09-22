resource "aws_route" "not_specified" { // aws_route_not_specified_target
  route_table_id         = "rtb-1234abcd"
  destination_cidr_block = "10.0.1.0/22"
}

resource "aws_route" "multiple_specified" { // aws_route_specified_multiple_targets
  route_table_id         = "rtb-1234abcd"
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

resource "aws_route" "not_specified2" { // aws_route_not_specified_target
  route_table_id         = "rtb-1234abcd"
  destination_cidr_block = "10.0.1.0/22"
}
