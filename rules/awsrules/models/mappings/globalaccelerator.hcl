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
