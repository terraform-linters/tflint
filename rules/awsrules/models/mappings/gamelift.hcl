import = "aws-sdk-go/models/apis/gamelift/2015-10-01/api-2.json"

mapping "aws_gamelift_alias" {
  name             = NonBlankAndLengthConstraintString
  description      = NonZeroAndMaxString
  routing_strategy = RoutingStrategy
}

mapping "aws_gamelift_build" {
  name             = NonZeroAndMaxString
  operating_system = OperatingSystem
  storage_location = S3Location
  version          = NonZeroAndMaxString
}

mapping "aws_gamelift_fleet" {
  build_id                           = BuildId
  ec2_instance_type                  = EC2InstanceType
  name                               = NonZeroAndMaxString
  description                        = NonZeroAndMaxString
  ec2_inbound_permission             = IpPermissionsList
  metric_groups                      = MetricGroupList
  new_game_session_protection_policy = ProtectionPolicy
  resource_creation_limit_policy     = ResourceCreationLimitPolicy
  runtime_configuration              = RuntimeConfiguration
}

mapping "aws_gamelift_game_session_queue" {
  name                  = GameSessionQueueName
  timeout_in_seconds    = WholeNumber
  destinations          = GameSessionQueueDestinationList
  player_latency_policy = PlayerLatencyPolicyList
}
