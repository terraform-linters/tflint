import = "aws-sdk-go/models/apis/monitoring/2010-08-01/api-2.json"

mapping "aws_cloudwatch_metric_alarm" {
  alarm_name                            = AlarmName
  comparison_operator                   = ComparisonOperator
  metric_name                           = MetricName
  namespace                             = Namespace
  statistic                             = Statistic
  alarm_description                     = AlarmDescription
  // TODO: Remove original aws_cloudwatch_metric_alarm_invalid_unit rule
  // unit                               = StandardUnit
  extended_statistic                    = ExtendedStatistic
  treat_missing_data                    = TreatMissingData
  evaluate_low_sample_count_percentiles = EvaluateLowSampleCountPercentile
}

test "aws_cloudwatch_metric_alarm" "comparison_operator" {
  ok = "GreaterThanOrEqualToThreshold"
  ng = "GreaterThanOrEqual"
}

test "aws_cloudwatch_metric_alarm" "namespace" {
  ok = "AWS/EC2"
  ng = ":EC2"
}

test "aws_cloudwatch_metric_alarm" "statistic" {
  ok = "Average"
  ng = "Median"
}

# TODO: Remove original aws_cloudwatch_metric_alarm_invalid_unit rule
/*
test "aws_cloudwatch_metric_alarm" "unit" {
  ok = "Gigabytes"
  ng = "GB"
}
*/

test "aws_cloudwatch_metric_alarm" "extended_statistic" {
  ok = "p100"
  ng = "p101"
}
