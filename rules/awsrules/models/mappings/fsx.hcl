import = "aws-sdk-go/models/apis/fsx/2018-03-01/api-2.json"

mapping "aws_fsx_lustre_file_system" {
  storage_capacity              = StorageCapacity
  subnet_ids                    = SubnetIds
  export_path                   = ArchivePath
  import_path                   = ArchivePath
  imported_file_chunk_size      = Megabytes
  security_group_ids            = SecurityGroupIds
  tags                          = Tags
  weekly_maintenance_start_time = WeeklyTime
}

mapping "aws_fsx_windows_file_system" {
  storage_capacity                  = StorageCapacity
  subnet_ids                        = SubnetIds
  throughput_capacity               = MegabytesPerSecond
  active_directory_id               = DirectoryId
  automatic_backup_retention_days   = AutomaticBackupRetentionDays
  copy_tags_to_backups              = Flag
  daily_automatic_backup_start_time = DailyTime
  kms_key_id                        = KmsKeyId
  security_group_ids                = SecurityGroupIds
  self_managed_active_directory     = SelfManagedActiveDirectoryConfiguration
  skip_final_backup                 = Flag
  tags                              = Tags
  weekly_maintenance_start_time     = WeeklyTime
}
