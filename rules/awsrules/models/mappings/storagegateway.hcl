import = "aws-sdk-go/models/apis/storagegateway/2013-06-30/api-2.json"

mapping "aws_storagegateway_cache" {
  disk_id     = DiskId
  gateway_arn = GatewayARN
}

mapping "aws_storagegateway_cached_iscsi_volume" {
  gateway_arn          = GatewayARN
  network_interface_id = NetworkInterfaceId
  target_name          = TargetName
  volume_size_in_bytes = long
  snapshot_id          = SnapshotId
  source_volume_arn    = VolumeARN
}

mapping "aws_storagegateway_gateway" {
  gateway_name                  = GatewayName
  gateway_timezone              = GatewayTimezone
  activation_key                = ActivationKey
  gateway_ip_address            = any
  gateway_type                  = GatewayType
  media_changer_type            = MediumChangerType
  smb_active_directory_settings = any
  smb_guest_password            = SMBGuestPassword
  tape_drive_type               = TapeDriveType
}

mapping "aws_storagegateway_nfs_file_share" {
  client_list             = FileShareClientList
  gateway_arn             = GatewayARN
  location_arn            = LocationARN
  role_arn                = Role
  default_storage_class   = StorageClass
  guess_mime_type_enabled = Boolean
  kms_encrypted           = Boolean
  kms_key_arn             = KMSKey
  nfs_file_share_defaults = NFSFileShareDefaults
  object_acl              = ObjectACL
  read_only               = Boolean
  requester_pays          = Boolean
  squash                  = Squash
}

mapping "aws_storagegateway_smb_file_share" {
  gateway_arn             = GatewayARN
  location_arn            = LocationARN
  role_arn                = Role
  authentication          = Authentication
  default_storage_class   = StorageClass
  guess_mime_type_enabled = Boolean
  invalid_user_list       = FileShareUserList
  kms_encrypted           = Boolean
  kms_key_arn             = KMSKey
  smb_file_share_defaults = NFSFileShareDefaults
  object_acl              = ObjectACL
  read_only               = Boolean
  requester_pays          = Boolean
  valid_user_list         = FileShareUserList
}

mapping "aws_storagegateway_upload_buffer" {
  disk_id     = DiskId
  gateway_arn = GatewayARN
}

mapping "aws_storagegateway_working_storage" {
  disk_id     = DiskId
  gateway_arn = GatewayARN
}
