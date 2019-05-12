import = "aws-sdk-go/models/apis/docdb/2014-10-31/api-2.json"

mapping "aws_docdb_cluster" {
  apply_immediately               = Boolean
  availability_zones              = AvailabilityZones
  backup_retention_period         = IntegerOptional
  cluster_identifier_prefix       = String
  cluster_identifier              = String
  db_subnet_group_name            = String
  db_cluster_parameter_group_name = String
  enabled_cloudwatch_logs_exports = LogTypeList
  engine_version                  = String
  engine                          = String
  final_snapshot_identifier       = String
  kms_key_id                      = String
  master_password                 = String
  master_username                 = String
  port                            = IntegerOptional
  preferred_backup_window         = String
  preferred_maintenance_window    = String
  skip_final_snapshot             = Boolean
  snapshot_identifier             = String
  storage_encrypted               = BooleanOptional
  tags                            = TagList
  vpc_security_group_ids          = VpcSecurityGroupIdList
}

mapping "aws_docdb_cluster_instance" {
  apply_immediately            = Boolean
  auto_minor_version_upgrade   = BooleanOptional
  availability_zone            = String
  cluster_identifier           = String
  engine                       = String
  identifier                   = String
  identifier_prefix            = String
  instance_class               = String
  preferred_maintenance_window = String
  promotion_tier               = IntegerOptional
  tags                         = TagList
}

mapping "aws_docdb_cluster_parameter_group" {
  name        = String
  name_prefix = String
  family      = String
  description = String
  parameter   = ParametersList
  tags        = TagList
}

mapping "aws_docdb_cluster_snapshot" {
  db_cluster_identifier          = String
  db_cluster_snapshot_identifier = String
}

mapping "aws_docdb_subnet_group" {
  name        = String
  name_prefix = String
  description = String
  subnet_ids  = SubnetIdentifierList
  tags        = TagList
}
