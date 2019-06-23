import = "aws-sdk-go/models/apis/dax/2017-04-19/api-2.json"

mapping "aws_dax_cluster" {
  cluster_name           = String
  iam_role_arn           = String
  node_type              = String
  replication_factor     = Integer
  availability_zones     = AvailabilityZoneList
  description            = String
  notification_topic_arn = String
  parameter_group_name   = String
  maintenance_window     = String
  security_group_ids     = SecurityGroupIdentifierList
  server_side_encryption = SSESpecification
  subnet_group_name      = String
  tags                   = TagList
}

mapping "aws_dax_parameter_group" {
  name        = String
  description = String
  parameters  = ParameterList
}

mapping "aws_dax_subnet_group" {
  name        = String
  description = String
  subnet_ids  = SubnetList
}
