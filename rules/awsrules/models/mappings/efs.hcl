import = "aws-sdk-go/models/apis/elasticfilesystem/2015-02-01/api-2.json"

mapping "aws_efs_file_system" {
  creation_token                  = CreationToken
  encrypted                       = Encrypted
  kms_key_id                      = KmsKeyId
  performance_mode                = PerformanceMode
  provisioned_throughput_in_mibps = ProvisionedThroughputInMibps
  tags                            = Tags
  throughput_mode                 = ThroughputMode
}

mapping "aws_efs_mount_target" {
  file_system_id  = FileSystemId
  subnet_id       = SubnetId
  ip_address      = IpAddress
  security_groups = SecurityGroups
}

test "aws_efs_file_system" "performance_mode" {
  ok = "generalPurpose"
  ng = "minIO"
}

test "aws_efs_file_system" "throughput_mode" {
  ok = "bursting"
  ng = "generalPurpose"
}
