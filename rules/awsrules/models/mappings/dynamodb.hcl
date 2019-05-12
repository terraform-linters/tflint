import = "aws-sdk-go/models/apis/dynamodb/2012-08-10/api-2.json"

mapping "aws_dynamodb_global_table" {
  name    = TableName
  replica = ReplicaList
}

mapping "aws_dynamodb_table" {
  name                   = TableName
  billing_mode           = BillingMode
  hash_key               = KeySchemaAttributeName
  range_key              = KeySchemaAttributeName
  write_capacity         = PositiveLongObject
  read_capacity          = PositiveLongObject
  attribute              = AttributeDefinitions
  local_secondary_index  = LocalSecondaryIndexList
  global_secondary_index = GlobalSecondaryIndexList
  stream_enabled         = StreamEnabled
  stream_view_type       = StreamViewType
  server_side_encryption = SSESpecification
  tags                   = TagList
  point_in_time_recovery = PointInTimeRecoverySpecification
}

mapping "aws_dynamodb_table_item" {
  table_name = TableName
  hash_key   = KeySchemaAttributeName
  range_key  = KeySchemaAttributeName
  item       = AttributeMap
}

test "aws_dynamodb_global_table" "name" {
  ok = "myTable"
  ng = "myTable@development"
}

test "aws_dynamodb_table" "billing_mode" {
  ok = "PROVISIONED"
  ng = "FLEXIBLE"
}

test "aws_dynamodb_table" "stream_view_type" {
  ok = "NEW_IMAGE"
  ng = "OLD_AND_NEW_IMAGE"
}
