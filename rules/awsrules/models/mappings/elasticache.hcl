import = "aws-sdk-go/models/apis/elasticache/2015-02-02/api-2.json"

mapping "aws_elasticache_cluster" {
  cluster_id                   = String
  replication_group_id         = String
  engine                       = String
  engine_version               = String
  maintenance_window           = String
  node_type                    = String
  num_cache_nodes              = IntegerOptional
  parameter_group_name         = String
  port                         = IntegerOptional
  subnet_group_name            = String
  security_group_names         = CacheSecurityGroupNameList
  security_group_ids           = SecurityGroupIdsList
  apply_immediately            = Boolean
  snapshot_arns                = SnapshotArnsList
  snapshot_name                = String
  snapshot_window              = String
  snapshot_retention_limit     = IntegerOptional
  notification_topic_arn       = String
  az_mode                      = AZMode
  availability_zone            = String
  preferred_availability_zones = PreferredAvailabilityZoneList
  tags                         = TagList
}

mapping "aws_elasticache_parameter_group" {
  name        = String
  family      = String
  description = String
  parameter   = ParametersList
}

mapping "aws_elasticache_replication_group" {
  replication_group_id          = String
  replication_group_description = String
  number_cache_clusters         = IntegerOptional
  node_type                     = String
  automatic_failover_enabled    = BooleanOptional
  auto_minor_version_upgrade    = BooleanOptional
  availability_zones            = AvailabilityZonesList
  engine                        = String
  at_rest_encryption_enabled    = BooleanOptional
  transit_encryption_enabled    = BooleanOptional
  auth_token                    = String
  engine_version                = String
  parameter_group_name          = String
  port                          = IntegerOptional
  subnet_group_name             = String
  security_group_names          = CacheSecurityGroupNameList
  security_group_ids            = SecurityGroupIdsList
  snapshot_arns                 = SnapshotArnsList
  snapshot_name                 = String
  maintenance_window            = String
  notification_topic_arn        = String
  snapshot_window               = String
  snapshot_retention_limit      = IntegerOptional
  apply_immediately             = Boolean
  tags                          = TagList
}

mapping "aws_elasticache_security_group" {
  name                 = String
  description          = String
  security_group_names = CacheSecurityGroupNameList
}

mapping "aws_elasticache_subnet_group" {
  name        = String
  description = String
  subnet_ids  = SubnetIdentifierList
}

test "aws_elasticache_cluster" "az_mode" {
  ok = "cross-az"
  ng = "multi-az"
}
