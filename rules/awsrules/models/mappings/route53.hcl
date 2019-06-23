import = "aws-sdk-go/models/apis/route53/2013-04-01/api-2.json"

mapping "aws_route53_delegation_set" {
  reference_name = Nonce
}

mapping "aws_route53_health_check" {
  reference_name                  = HealthCheckNonce
  fqdn                            = FullyQualifiedDomainName
  ip_address                      = IPAddress
  port                            = Port
  type                            = HealthCheckType
  failure_threshold               = FailureThreshold
  request_interval                = RequestInterval
  resource_path                   = ResourcePath
  search_string                   = SearchString
  measure_latency                 = MeasureLatency
  invert_healthcheck              = Inverted
  enable_sni                      = EnableSNI
  child_healthchecks              = ChildHealthCheckList
  child_health_threshold          = HealthThreshold
  cloudwatch_alarm_name           = AlarmName
  cloudwatch_alarm_region         = CloudWatchRegion
  insufficient_data_health_status = InsufficientDataHealthStatus
  regions                         = HealthCheckRegionList
  tags                            = TagList
}

mapping "aws_route53_query_log" {
  cloudwatch_log_group_arn = CloudWatchLogsLogGroupArn
  zone_id                  = ResourceId
}

mapping "aws_route53_record" {
  zone_id                          = ResourceId
  name                             = DNSName
  type                             = RRType
  ttl                              = TTL
  records                          = ResourceRecords
  set_identifier                   = ResourceRecordSetIdentifier
  health_check_id                  = HealthCheckId
  alias                            = AliasTarget
  failover_routing_policy          = any
  geolocation_routing_policy       = GeoLocation
  latency_routing_policy           = any
  weighted_routing_policy          = any
  multivalue_answer_routing_policy = ResourceRecordSetMultiValueAnswer
  allow_overwrite                  = any
}

mapping "aws_route53_zone" {
  name              = DNSName
  comment           = ResourceDescription
  delegation_set_id = ResourceId
  force_destroy     = any
  tags              = TagList
  vpc               = VPC
}

mapping "aws_route53_zone_association" {
  zone_id    = ResourceId
  vpc_id     = VPCId
  vpc_region = VPCRegion
}
