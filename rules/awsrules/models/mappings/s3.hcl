import = "aws-sdk-go/models/apis/s3/2006-03-01/api-2.json"

mapping "aws_s3_account_public_access_block" {
  account_id              = AccountId
  block_public_acls       = Setting
  block_public_policy     = Setting
  ignore_public_acls      = Setting
  restrict_public_buckets = Setting
}

mapping "aws_s3_bucket" {
  bucket                               = BucketName
  bucket_prefix                        = any
  acl                                  = BucketCannedACL
  policy                               = Policy
  tags                                 = TagSet
  force_destroy                        = any
  website                              = WebsiteConfiguration
  cors_rule                            = CORSRules
  versioning                           = VersioningConfiguration
  logging                              = LoggingEnabled
  lifecycle_rule                       = LifecycleRules
  acceleration_status                  = BucketAccelerateStatus
  region                               = BucketLocationConstraint
  request_payer                        = Payer
  replication_configuration            = ReplicationConfiguration
  server_side_encryption_configuration = ServerSideEncryptionConfiguration
  object_lock_configuration            = ObjectLockConfiguration
}

mapping "aws_s3_bucket_inventory" {
  bucket                   = BucketName
  name                     = InventoryId
  included_object_versions = InventoryIncludedObjectVersions
  schedule                 = InventorySchedule
  destination              = InventoryDestination
  enabled                  = IsEnabled
  filter                   = InventoryFilter
  optional_fields          = InventoryOptionalFields
}

mapping "aws_s3_bucket_metric" {
  bucket = BucketName
  name   = MetricsId
  filter = MetricsFilter
}

mapping "aws_s3_bucket_notification" {
  bucket          = BucketName
  topic           = TopicConfiguration
  queue           = QueueConfiguration
  lambda_function = LambdaFunctionConfiguration
}

mapping "aws_s3_bucket_object" {
  bucket                 = BucketName
  key                    = ObjectKey
  source                 = any
  content                = any
  content_base64         = any
  acl                    = ObjectCannedACL
  cache_control          = CacheControl
  content_disposition    = ContentDisposition
  content_encoding       = ContentEncoding
  content_language       = ContentLanguage
  content_type           = ContentType
  website_redirect       = WebsiteRedirectLocation
  storage_class          = StorageClass
  etag                   = ETag
  server_side_encryption = ServerSideEncryption
  kms_key_id             = SSEKMSKeyId
  tags                   = TagSet
}

mapping "aws_s3_bucket_policy" {
  bucket = BucketName
  policy = Policy
}

mapping "aws_s3_bucket_public_access_block" {
  bucket                  = BucketName
  block_public_acls       = Setting
  block_public_policy     = Setting
  ignore_public_acls      = Setting
  restrict_public_buckets = Setting
}
