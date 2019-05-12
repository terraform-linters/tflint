import = "aws-sdk-go/models/apis/rds/2014-10-31/api-2.json"

mapping "aws_db_cluster_snapshot" {
  db_cluster_identifier          = String
  db_cluster_snapshot_identifier = String
}

mapping "aws_db_event_subscription" {
  name             = String
  name_prefix      = any
  sns_topic        = String
  source_ids       = SourceIdsList
  source_type      = String
  event_categories = EventCategoriesList
  enabled          = BooleanOptional
  tags             = TagList
}

mapping "aws_db_instance" {
  allocated_storage                     = IntegerOptional
  allow_major_version_upgrade           = Boolean
  apply_immediately                     = Boolean
  auto_minor_version_upgrade            = BooleanOptional
  availability_zone                     = String
  backup_retention_period               = IntegerOptional
  backup_window                         = String
  character_set_name                    = String
  copy_tags_to_snapshot                 = BooleanOptional
  db_subnet_group_name                  = String
  deletion_protection                   = BooleanOptional
  domain                                = String
  domain_iam_role_name                  = String
  enabled_cloudwatch_logs_exports       = LogTypeList
  engine                                = String
  engine_version                        = String
  final_snapshot_identifier             = String
  iam_database_authentication_enabled   = BooleanOptional
  identifier                            = String
  identifier_prefix                     = any
  instance_class                        = String
  iops                                  = IntegerOptional
  kms_key_id                            = String
  license_model                         = String
  maintenance_window                    = String
  monitoring_interval                   = IntegerOptional
  monitoring_role_arn                   = String
  multi_az                              = BooleanOptional
  name                                  = String
  option_group_name                     = String
  parameter_group_name                  = String
  password                              = String
  port                                  = IntegerOptional
  publicly_accessible                   = BooleanOptional
  replicate_source_db                   = any
  security_group_names                  = DBSecurityGroupNameList
  skip_final_snapshot                   = Boolean
  snapshot_identifier                   = String
  storage_encrypted                     = BooleanOptional
  storage_type                          = String
  tags                                  = TagList
  timezone                              = String
  username                              = String
  vpc_security_group_ids                = VpcSecurityGroupIdList
  s3_import                             = any
  performance_insights_enabled          = BooleanOptional
  performance_insights_kms_key_id       = String
  performance_insights_retention_period = IntegerOptional
}

mapping "aws_db_instance_role_association" {
  db_instance_identifier = String
  feature_name           = String
  role_arn               = String
}

mapping "aws_db_option_group" {
  name                     = String
  name_prefix              = any
  option_group_description = String
  engine_name              = String
  major_engine_version     = String
  option                   = OptionsList
  tags                     = TagList
}

mapping "aws_db_parameter_group" {
  name        = String
  name_prefix = any
  family      = String
  description = String
  parameter   = ParametersList
  tags        = TagList
}

mapping "aws_db_security_group" {
  name        = String
  description = String
  ingress     = AuthorizeDBSecurityGroupIngressMessage
  tags        = TagList
}

mapping "aws_db_snapshot" {
  db_instance_identifier = String
  db_snapshot_identifier = String
  tags                   = TagList
}

mapping "aws_db_subnet_group" {
  name        = String
  name_prefix = any
  description = String
  subnet_ids  = SubnetIdentifierList
  tags        = TagList
}

mapping "aws_rds_cluster" {
  cluster_identifier                  = String
  cluster_identifier_prefix           = any
  copy_tags_to_snapshot               = BooleanOptional
  database_name                       = String
  deletion_protection                 = BooleanOptional
  master_password                     = String
  master_username                     = String
  final_snapshot_identifier           = String
  skip_final_snapshot                 = Boolean
  availability_zones                  = AvailabilityZones
  backtrack_window                    = LongOptional
  backup_retention_period             = IntegerOptional
  preferred_backup_window             = String
  preferred_maintenance_window        = String
  port                                = IntegerOptional
  vpc_security_group_ids              = VpcSecurityGroupIdList
  snapshot_identifier                 = String
  global_cluster_identifier           = String
  storage_encrypted                   = BooleanOptional
  replication_source_identifier       = String
  apply_immediately                   = Boolean
  db_subnet_group_name                = String
  db_cluster_parameter_group_name     = String
  kms_key_id                          = String
  iam_roles                           = any
  iam_database_authentication_enabled = BooleanOptional
  engine                              = String
  engine_mode                         = String
  engine_version                      = String
  source_region                       = String
  enabled_cloudwatch_logs_exports     = LogTypeList
  scaling_configuration               = ScalingConfiguration
  tags                                = TagList
}

mapping "aws_rds_cluster_endpoint" {
  cluster_identifier          = String
  cluster_endpoint_identifier = String
  custom_endpoint_type        = String
  static_members              = StringList
  excluded_members            = StringList
}

mapping "aws_rds_cluster_instance" {
  identifier                      = String
  identifier_prefix               = any
  cluster_identifier              = String
  engine                          = String
  engine_version                  = String
  instance_class                  = String
  publicly_accessible             = BooleanOptional
  db_subnet_group_name            = String
  db_parameter_group_name         = String
  apply_immediately               = Boolean
  monitoring_role_arn             = String
  monitoring_interval             = IntegerOptional
  promotion_tier                  = IntegerOptional
  availability_zone               = String
  preferred_backup_window         = String
  preferred_maintenance_window    = String
  auto_minor_version_upgrade      = BooleanOptional
  performance_insights_enabled    = BooleanOptional
  performance_insights_kms_key_id = String
  copy_tags_to_snapshot           = BooleanOptional
  tags                            = TagList
}

mapping "aws_rds_cluster_parameter_group" {
  name        = String
  name_prefix = any
  family      = String
  description = String
  parameter   = ParametersList
  tags        = TagList
}

mapping "aws_rds_global_cluster" {
  database_name       = String
  deletion_protection = BooleanOptional
  engine              = String
  engine_version      = String
  storage_encrypted   = BooleanOptional
}
