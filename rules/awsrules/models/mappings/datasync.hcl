import = "aws-sdk-go/models/apis/datasync/2018-11-09/api-2.json"

mapping "aws_datasync_agent" {
  name           = TagValue
  activation_key = ActivationKey
  tags           = TagList
}

mapping "aws_datasync_location_efs" {
  ec2_config          = Ec2Config
  efs_file_system_arn = EfsFilesystemArn
  subdirectory        = Subdirectory
  tags                = TagList
}

mapping "aws_datasync_location_nfs" {
  on_prem_config  = OnPremConfig
  server_hostname = ServerHostname
  subdirectory    = Subdirectory
  tags            = TagList
}

mapping "aws_datasync_location_s3" {
  s3_bucket_arn = S3BucketArn
  s3_config     = S3Config
  subdirectory  = Subdirectory
  tags          = TagList
}

mapping "aws_datasync_task" {
  destination_location_arn = LocationArn
  source_location_arn      = LocationArn
  cloudwatch_log_group_arn = LogGroupArn
  name                     = TagValue
  options                  = Options
  tags                     = TagList
}

test "aws_datasync_agent" "name" {
  ok = "example"
  ng = "example^example"
}

test "aws_datasync_agent" "activation_key" {
  ok = "F0EFT-7FPPR-GG7MC-3I9R3-27DOH"
  ng = "F0EFT7FPPRGG7MC3I9R327DOH"
}

test "aws_datasync_location_efs" "efs_file_system_arn" {
  ok = "arn:aws:elasticfilesystem:us-east-1:123456789012:file-system/fs-12345678"
  ng = "arn:aws:eks:us-east-1:123456789012:cluster/my-cluster"
}

test "aws_datasync_location_efs" "subdirectory" {
  ok = "foo"
  ng = "bar\t"
}

test "aws_datasync_location_nfs" "server_hostname" {
  ok = "nfs.example.com"
  ng = "nfs^example^com"
}

test "aws_datasync_location_nfs" "subdirectory" {
  ok = "/exported/path"
  ng = "/exported^path"
}

test "aws_datasync_location_s3" "s3_bucket_arn" {
  ok = "arn:aws:s3:::my_corporate_bucket"
  ng = "arn:aws:eks:us-east-1:123456789012:cluster/my-cluster"
}

test "aws_datasync_task" "cloudwatch_log_group_arn" {
  ok = "arn:aws:logs:us-east-1:123456789012:log-group:my-log-group"
  ng = "arn:aws:s3:::my_corporate_bucket"
}

test "aws_datasync_task" "source_location_arn" {
  ok = "arn:aws:datasync:us-east-2:111222333444:location/loc-07db7abfc326c50fb"
  ng = "arn:aws:datasync:us-east-2:111222333444:task/task-08de6e6697796f026"
}
