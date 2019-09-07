import = "aws-sdk-go/models/apis/redshift/2012-12-01/api-2.json"

mapping "aws_redshift_cluster" {
  cluster_identifier                  = String
  database_name                       = String
  node_type                           = String
  cluster_type                        = String
  master_password                     = String
  master_username                     = String
  cluster_security_groups             = ClusterSecurityGroupNameList
  vpc_security_group_ids              = VpcSecurityGroupIdList
  cluster_subnet_group_name           = String
  availability_zone                   = String
  preferred_maintenance_window        = String
  cluster_parameter_group_name        = String
  automated_snapshot_retention_period = IntegerOptional
  port                                = IntegerOptional
  cluster_version                     = String
  allow_version_upgrade               = BooleanOptional
  number_of_nodes                     = IntegerOptional
  publicly_accessible                 = BooleanOptional
  encrypted                           = BooleanOptional
  enhanced_vpc_routing                = Boolean
  kms_key_id                          = String
  elastic_ip                          = String
  skip_final_snapshot                 = Boolean
  final_snapshot_identifier           = String
  snapshot_identifier                 = String
  snapshot_cluster_identifier         = String
  owner_account                       = String
  iam_roles                           = IamRoleArnList
  logging                             = LoggingStatus
  snapshot_copy                       = EnableSnapshotCopyMessage
  tags                                = TagList
}

mapping "aws_redshift_event_subscription" {
  name             = String
  sns_topic_arn    = String
  source_ids       = SourceIdsList
  source_type      = String
  severity         = String
  event_categories = EventCategoriesList
  enabled          = Boolean
  tags             = TagList
}

mapping "aws_redshift_parameter_group" {
  name        = String
  family      = String
  description = String
  parameter   = ParametersList
  tags        = TagList
}

mapping "aws_redshift_security_group" {
  name        = String
  description = String
  ingress     = AuthorizeClusterSecurityGroupIngressMessage
}

mapping "aws_redshift_snapshot_copy_grant" {
  snapshot_copy_grant_name = String
  kms_key_id               = String
  tags                     = TagList
}

mapping "aws_redshift_snapshot_schedule" {
  identifier        = String
  identifier_prefix = String
  description       = String
  definitions       = ScheduleDefinitionList
  force_destroy     = any
  tags              = TagList
}

mapping "aws_redshift_snapshot_schedule_association" {
  cluster_identifier  = String
  schedule_identifier = String
}

mapping "aws_redshift_subnet_group" {
  name        = String
  description = String
  subnet_ids  = SubnetIdentifierList
  tags        = TagList
}
