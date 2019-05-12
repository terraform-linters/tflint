import = "aws-sdk-go/models/apis/dms/2016-01-01/api-2.json"

mapping "aws_dms_certificate" {
  certificate_id = String
  certificate_pem = String
  certificate_wallet = CertificateWallet
}

mapping "aws_dms_endpoint" {
  certificate_arn             = String
  database_name               = String
  endpoint_id                 = String
  endpoint_type               = ReplicationEndpointTypeValue
  engine_name                 = String
  extra_connection_attributes = String
  kms_key_arn                 = String
  password                    = SecretString
  port                        = IntegerOptional
  server_name                 = String
  ssl_mode                    = DmsSslModeValue
  tags                        = TagList
  username                    = String
  service_access_role         = String
  mongodb_settings            = MongoDbSettings
  s3_settings                 = S3Settings
}

mapping "aws_dms_replication_instance" {
  allocated_storage            = IntegerOptional
  apply_immediately            = Boolean
  auto_minor_version_upgrade   = BooleanOptional
  availability_zone            = String
  engine_version               = String
  kms_key_arn                  = String
  multi_az                     = BooleanOptional
  preferred_maintenance_window = String
  publicly_accessible          = BooleanOptional
  replication_instance_class   = String
  replication_instance_id      = String
  replication_subnet_group_id  = String
  tags                         = TagList
  vpc_security_group_ids       = VpcSecurityGroupIdList
}

mapping "aws_dms_replication_subnet_group" {
  replication_subnet_group_description = String
  replication_subnet_group_id          = String
  subnet_ids                           = SubnetIdentifierList
  tags                                 = TagList
}

mapping "aws_dms_replication_task" {
  cdc_start_time            = TStamp
  migration_type            = MigrationTypeValue
  replication_instance_arn  = String
  replication_task_id       = String
  replication_task_settings = String
  source_endpoint_arn       = String
  table_mappings            = String
  tags                      = TagList
  target_endpoint_arn       = String
}

test "aws_dms_endpoint" "endpoint_type" {
  ok = "source"
  ng = "resource"
}

test "aws_dms_endpoint" "ssl_mode" {
  ok = "require"
  ng = "verify-require"
}

test "aws_dms_replication_task" "migration_type" {
  ok = "full-load"
  ng = "partial-load"
}
