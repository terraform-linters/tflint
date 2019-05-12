import = "aws-sdk-go/models/apis/ec2/2016-11-15/api-2.json"

mapping "aws_customer_gateway" {
  bgp_asn    = Integer
  ip_address = String
  type       = GatewayType
  tags       = TagList
}

mapping "aws_default_network_acl" {
  default_network_acl_id = String
  subnet_ids             = ValueStringList
  ingress                = any
  egress                 = any
  tags                   = TagList
}

mapping "aws_default_route_table" {
  default_route_table_id = String
  route                  = RouteList
  tags                   = TagList
  propagating_vgws       = RouteTableAssociationList
}

mapping "aws_default_security_group" {
  ingress = any
  egress  = any
  vpc_id  = String
  tags    = TagList
}

mapping "aws_default_subnet" {
  map_public_ip_on_launch = any
  tags                    = TagList
}

mapping "aws_default_vpc" {
  enable_dns_support   = any
  enable_dns_hostnames = any
  enable_classiclink   = any
  tags                 = TagList
}

mapping "aws_default_vpc_dhcp_options" {
  netbios_name_servers = any
  netbios_node_type    = any
  tags                 = TagList
}

mapping "aws_egress_only_internet_gateway" {
  vpc_id = String
}

mapping "aws_flow_log" {
  traffic_type         = TrafficType
  eni_id               = String
  iam_role_arn         = String
  log_destination_type = LogDestinationType
  log_destination      = String
  log_group_name       = String
  subnet_id            = String
  vpc_id               = String
}

mapping "aws_internet_gateway" {
  vpc_id = String
  tags   = TagList
}

mapping "aws_main_route_table_association" {
  vpc_id         = String
  route_table_id = String
}

mapping "aws_nat_gateway" {
  allocation_id = String
  subnet_id     = String
  tags          = TagList
}

mapping "aws_network_acl" {
  vpc_id     = String
  subnet_ids = ValueStringList
  ingress    = any
  egress     = any
  tags       = TagList
}

mapping "aws_network_acl_rule" {
  network_acl_id  = String
  rule_number     = Integer
  egress          = Boolean
  protocol        = String
  rule_action     = RuleAction
  cidr_block      = String
  ipv6_cidr_block = String
  from_port       = Integer
  to_port         = Integer
  icmp_type       = Integer
  icmp_code       = Integer
}

mapping "aws_network_interface" {
  subnet_id         = String
  description       = String
  private_ips       = PrivateIpAddressSpecificationList
  private_ips_count = Integer
  security_groups   = SecurityGroupIdStringList
  attachment        = NetworkInterfaceAttachment
  source_dest_check = AttributeBooleanValue
  tags              = TagList
}

mapping "aws_network_interface_attachment" {
  instance_id          = String
  network_interface_id = String
  device_index         = Integer
}

mapping "aws_route" {
  route_table_id              = String
  destination_cidr_block      = String
  destination_ipv6_cidr_block = String
  egress_only_gateway_id      = String
  gateway_id                  = String
  instance_id                 = String
  nat_gateway_id              = String
  network_interface_id        = String
  transit_gateway_id          = String
  vpc_peering_connection_id   = String
}

mapping "aws_route_table" {
  vpc_id           = String
  route            = RouteList
  tags             = TagList
  propagating_vgws = RouteTableAssociationList
}

mapping "aws_route_table_association" {
  subnet_id      = String
  route_table_id = String
}

mapping "aws_security_group" {
  name                   = String
  name_prefix            = any
  description            = String
  ingress                = any
  egress                 = any
  revoke_rules_on_delete = any
  vpc_id                 = String
  tags                   = TagList
}

mapping "aws_network_interface_sg_attachment" {
  security_group_id    = String
  network_interface_id = String
}

mapping "aws_security_group_rule" {
  type                     = any
  cidr_blocks              = any
  ipv6_cidr_blocks         = any
  prefix_list_ids          = any
  from_port                = Integer
  protocol                 = String
  security_group_id        = String
  source_security_group_id = String
  self                     = any
  to_port                  = Integer
  description              = any
}

mapping "aws_subnet" {
  availability_zone               = String
  availability_zone_id            = String
  cidr_block                      = String
  ipv6_cidr_block                 = String
  map_public_ip_on_launch         = any
  assign_ipv6_address_on_creation = any
  vpc_id                          = String
  tags                            = TagList
}

mapping "aws_vpc" {
  cidr_block                       = String
  instance_tenancy                 = Tenancy
  enable_dns_support               = any
  enable_dns_hostnames             = any
  enable_classiclink               = any
  enable_classiclink_dns_support   = any
  assign_generated_ipv6_cidr_block = any
  tags                             = TagList
}

mapping "aws_vpc_dhcp_options" {
  domain_name          = any
  domain_name_servers  = any
  ntp_servers          = any
  netbios_name_servers = any
  netbios_node_type    = any
  tags                 = TagList
}

mapping "aws_vpc_dhcp_options_association" {
  vpc_id          = String
  dhcp_options_id = String
}

mapping "aws_vpc_endpoint" {
  service_name        = String
  vpc_id              = String
  auto_accept         = any
  policy              = String
  private_dns_enabled = Boolean
  route_table_ids     = ValueStringList
  subnet_ids          = ValueStringList
  security_group_ids  = ValueStringList
  tags                = TagList
  vpc_endpoint_type   = VpcEndpointType
}

mapping "aws_vpc_endpoint_connection_notification" {
  vpc_endpoint_service_id     = String
  vpc_endpoint_id             = String
  connection_notification_arn = String
  connection_events           = ValueStringList
}

mapping "aws_vpc_endpoint_route_table_association" {
  route_table_id  = String
  vpc_endpoint_id = String
}

mapping "aws_vpc_endpoint_service" {
  acceptance_required        = Boolean
  network_load_balancer_arns = ValueStringList
  allowed_principals         = any
  tags                       = TagList
}

mapping "aws_vpc_endpoint_service_allowed_principal" {
  vpc_endpoint_service_id = String
  principal_arn           = String
}

mapping "aws_vpc_endpoint_subnet_association" {
  vpc_endpoint_id = String
  subnet_id       = String
}

mapping "aws_vpc_ipv4_cidr_block_association" {
  cidr_block = String
  vpc_id     = String
}

mapping "aws_vpc_peering_connection" {
  peer_owner_id = String
  peer_vpc_id   = String
  vpc_id        = String
  auto_accept   = any
  peer_region   = String
  accepter      = PeeringConnectionOptionsRequest
  requester     = PeeringConnectionOptionsRequest
  tags          = TagList
}

mapping "aws_vpc_peering_connection_accepter" {
  vpc_peering_connection_id = String
  auto_accept               = any
  tags                      = TagList
}

mapping "aws_vpc_peering_connection_options" {
  vpc_peering_connection_id = String
  accepter                  = PeeringConnectionOptionsRequest
  requester                 = PeeringConnectionOptionsRequest
}

mapping "aws_vpn_connection" {
  customer_gateway_id   = String
  type                  = String
  transit_gateway_id    = String
  vpn_gateway_id        = String
  static_routes_only    = any
  tags                  = TagList
  tunnel1_inside_cidr   = any
  tunnel2_inside_cidr   = any
  tunnel1_preshared_key = any
  tunnel2_preshared_key = any
}

mapping "aws_vpn_connection_route" {
  destination_cidr_block = String
  vpn_connection_id      = String
}

mapping "aws_vpn_gateway" {
  vpc_id            = String
  availability_zone = String
  tags              = TagList
  amazon_side_asn   = Long
}

mapping "aws_vpn_gateway_attachment" {
  vpc_id         = String
  vpn_gateway_id = String
}

mapping "aws_vpn_gateway_route_propagation" {
  vpn_gateway_id = String
  route_table_id = String
}
