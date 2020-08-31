import = "aws-sdk-go/models/apis/cur/2017-01-06/api-2.json"

mapping "aws_cur_report_definition" {
  report_name = ReportName
  time_unit   = TimeUnit
  format      = ReportFormat
  compression = CompressionFormat
  s3_bucket   = S3Bucket
  s3_prefix   = S3Prefix
  s3_region   = AWSRegion
}

test "aws_cur_report_definition" "report_name" {
  ok = "example-cur-report-definition"
  ng = "example/cur-report-definition"
}

test "aws_cur_report_definition" "time_unit" {
  ok = "HOURLY"
  ng = "FORNIGHTLY"
}

test "aws_cur_report_definition" "format" {
  ok = "textORcsv"
  ng = "textORjson"
}

test "aws_cur_report_definition" "compression" {
  ok = "ZIP"
  ng = "TAR"
}

test "aws_cur_report_definition" "s3_region" {
  ok = "us-east-1"
  ng = "us-gov-east-1"
}
