# AWS CloudWatch Metric Alarm Invalid Unit
This issue reported if you specify invalid unit. This issue type is ERROR.

## Example
```
resource "aws_cloudwatch_metric_alarm" "rds-writer-memory" {
  alarm_name                = "terraform-test-foobar5"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  unit                      = "percent" // valid unit is "Percent"
  alarm_description         = "This metric monitor ec2 cpu utilization"
  insufficient_data_actions = []
}
```

The following is the execution result of TFLint:

```
template.tf
	ERROR:10 "percent" is invalid unit.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
Only the followings are supported for CloudWatch alarm unit. If an invalid unit is specified, an error will occur at `terraform apply`.

- Seconds
- Microseconds
- Milliseconds
- Bytes
- Kilobytes
- Megabytes
- Gigabytes
- Terabytes
- Bits
- Kilobits
- Megabits
- Gigabits
- Terabits
- Percent
- Count
- Bytes/Second
- Kilobytes/Second
- Megabytes/Second
- Gigabytes/Second
- Terabytes/Second
- Bits/Second
- Kilobits/Second
- Megabits/Second
- Gigabits/Second
- Terabits/Second
- Count/Second
- None

## How To Fix
Check the unit and select a valid unit again.
