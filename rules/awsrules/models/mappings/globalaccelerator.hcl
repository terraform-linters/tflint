import = "aws-sdk-go/models/apis/globalaccelerator/2018-08-08/api-2.json"

mapping "aws_globalaccelerator_accelerator" {
  name            = GenericString
  ip_address_type = IpAddressType
  enabled         = GenericBoolean
  attributes      = AcceleratorAttributes
}

mapping "aws_globalaccelerator_listener" {
  accelerator_arn = GenericString
  client_affinity = ClientAffinity
  protocol        = Protocol
  port_range      = PortRanges
}

mapping "aws_globalaccelerator_endpoint_group" {
  listener_arn                  = GenericString
  health_check_interval_seconds = HealthCheckIntervalSeconds
  health_check_path             = GenericString
  health_check_port             = HealthCheckPort
  health_check_protocol         = HealthCheckProtocol
  threshold_count               = ThresholdCount
  traffic_dial_percentage       = TrafficDialPercentage
  endpoint_configuration        = EndpointConfigurations
}
