import = "aws-sdk-go/models/apis/neptune/2014-10-31/api-2.json"

mapping "aws_neptune_parameter_group" {
  name        = String
  family      = String
  description = String
  parameter   = ParametersList
  tags        = TagList
}

mapping "aws_neptune_subnet_group" {
  name        = String
  name_prefix = any
  description = String
  subnet_ids  = SubnetIdentifierList
  tags        = TagList
}

mapping "aws_neptune_cluster_parameter_group" {
  name        = String
  name_prefix = any
  family      = String
  description = String
  parameter   = ParametersList
  tags        = TagList
}

mapping "aws_neptune_cluster" {
  apply_immediately                    = any
  availability_zones                   = AvailabilityZones
  backup_retention_period              = IntegerOptional
  cluster_identifier                   = String
  cluster_identifier_prefix            = any
  engine                               = String
  engine_version                       = String
  final_snapshot_identifier            = String
  iam_roles                            = any
  iam_database_authentication_enabled  = BooleanOptional
  kms_key_arn                          = String
  neptune_subnet_group_name            = String
  neptune_cluster_parameter_group_name = String
  preferred_backup_window              = String
  preferred_maintenance_window         = String
  port                                 = IntegerOptional
  replication_source_identifier        = String
  skip_final_snapshot                  = Boolean
  snapshot_identifier                  = String
  storage_encrypted                    = BooleanOptional
  tags                                 = TagList
  vpc_security_group_ids               = VpcSecurityGroupIdList
}

mapping "aws_neptune_cluster_instance" {
  apply_immediately            = any
  auto_minor_version_upgrade   = BooleanOptional
  availability_zone            = String
  cluster_identifier           = String
  engine                       = String
  engine_version               = String
  identifier                   = String
  identifier_prefix            = any
  instance_class               = String
  neptune_subnet_group_name    = String
  neptune_parameter_group_name = String
  port                         = IntegerOptional
  preferred_backup_window      = String
  preferred_maintenance_window = String
  promotion_tier               = IntegerOptional
  publicly_accessible          = BooleanOptional
  tags                         = TagList
}

mapping "aws_neptune_cluster_snapshot" {
  db_cluster_identifier          = String
  db_cluster_snapshot_identifier = String
}

mapping "aws_neptune_event_subscription" {
  enabled          = BooleanOptional
  event_categories = EventCategoriesList
  name             = String
  name_prefix      = any
  sns_topic_arn    = String
  source_ids       = SourceIdsList
  source_type      = String
  tags             = TagList
}
