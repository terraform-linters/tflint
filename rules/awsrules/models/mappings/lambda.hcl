import = "aws-sdk-go/models/apis/lambda/2015-03-31/api-2.json"

mapping "aws_lambda_alias" {
  name             = any // Alias
  description      = Description
  function_name    = FunctionName
  function_version = Version
  routing_config   = AliasRoutingConfiguration
}

mapping "aws_lambda_event_source_mapping" {
  batch_size                  = BatchSize
  event_source_arn            = Arn
  enabled                     = Enabled
  function_name               = FunctionName
  starting_position           = EventSourcePosition
  starting_position_timestamp = Date
}

mapping "aws_lambda_function" {
  filename                       = any
  s3_bucket                      = any // S3Bucket
  s3_key                         = S3Key
  s3_object_version              = S3ObjectVersion
  function_name                  = FunctionName
  dead_letter_config             = DeadLetterConfig
  tracing_config                 = TracingConfig
  handler                        = Handler
  role                           = RoleArn
  description                    = Description
  layers                         = LayerList
  memory_size                    = MemorySize
  runtime                        = Runtime
  timeout                        = Timeout
  reserved_concurrent_executions = ReservedConcurrentExecutions
  publish                        = Boolean
  vpc_config                     = VpcConfig
  environment                    = Environment
  kms_key_arn                    = KMSKeyArn
  source_code_hash               = any
  tags                           = Tags
}

mapping "aws_lambda_layer_version" {
  layer_name          = LayerName
  filename            = any
  s3_bucket           = any // S3Bucket
  s3_key              = S3Key
  s3_object_version   = S3ObjectVersion
  compatible_runtimes = CompatibleRuntimes
  description         = Description
  license_info        = LicenseInfo
  source_code_hash    = any
}

mapping "aws_lambda_permission" {
  action              = Action
  event_source_token  = EventSourceToken
  function_name       = FunctionName
  principal           = Principal
  qualifier           = Qualifier
  source_account      = SourceOwner
  source_arn          = Arn
  statement_id        = StatementId
  statement_id_prefix = any
}
