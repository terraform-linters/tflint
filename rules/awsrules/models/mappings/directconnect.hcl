import = "aws-sdk-go/models/apis/directconnect/2012-10-25/api-2.json"

mapping "aws_dx_bgp_peer" {
  address_family       = AddressFamily
  bgp_asn              = ASN
  virtual_interface_id = VirtualInterfaceId
  amazon_address       = AmazonAddress
  bgp_auth_key         = BGPAuthKey
  customer_address     = CustomerAddress
}

mapping "aws_dx_connection" {
  name      = ConnectionName
  bandwidth = Bandwidth
  location  = LocationCode
  tags      = TagList
}

mapping "aws_dx_connection_association" {
  connection_id = ConnectionId
  lag_id        = LagId
}

mapping "aws_dx_gateway" {
  name            = DirectConnectGatewayName
  amazon_side_asn = LongAsn
}

mapping "aws_dx_gateway_association" {
  dx_gateway_id                       = DirectConnectGatewayId
  associated_gateway_id               = AssociatedGatewayId
  vpn_gateway_id                      = AssociatedGatewayId
  associated_gateway_owner_account_id = OwnerAccount
  proposal_id                         = DirectConnectGatewayAssociationProposalId
  allowed_prefixes                    = RouteFilterPrefixList
}

mapping "aws_dx_gateway_association_proposal" {
  dx_gateway_id               = DirectConnectGatewayId
  dx_gateway_owner_account_id = OwnerAccount
  associated_gateway_id       = AssociatedGatewayId
  vpn_gateway_id              = AssociatedGatewayId
  allowed_prefixes            = RouteFilterPrefixList
}

mapping "aws_dx_hosted_private_virtual_interface" {
  address_family   = AddressFamily
  bgp_asn          = ASN
  connection_id    = ConnectionId
  name             = VirtualInterfaceName
  owner_account_id = OwnerAccount
  vlan             = VLAN
  amazon_address   = AmazonAddress
  mtu              = MTU
  bgp_auth_key     = BGPAuthKey
  customer_address = CustomerAddress
}

mapping "aws_dx_hosted_private_virtual_interface_accepter" {
  virtual_interface_id = VirtualInterfaceId
  dx_gateway_id        = DirectConnectGatewayId
  tags                 = TagList
  vpn_gateway_id       = AssociatedGatewayId
}

mapping "aws_dx_hosted_public_virtual_interface" {
  address_family        = AddressFamily
  bgp_asn               = ASN
  connection_id         = ConnectionId
  name                  = VirtualInterfaceName
  owner_account_id      = OwnerAccount
  route_filter_prefixes = RouteFilterPrefixList
  vlan                  = VLAN
  amazon_address        = AmazonAddress
  bgp_auth_key          = BGPAuthKey
  customer_address      = CustomerAddress
}

mapping "aws_dx_hosted_public_virtual_interface_accepter" {
  virtual_interface_id = VirtualInterfaceId
  tags                 = TagList
}

mapping "aws_dx_lag" {
  name                  = LagName
  connections_bandwidth = Bandwidth
  location              = LocationCode
  tags                  = TagList
}

mapping "aws_dx_private_virtual_interface" {
  address_family   = AddressFamily
  bgp_asn          = ASN
  connection_id    = ConnectionId
  name             = VirtualInterfaceName
  vlan             = VLAN
  amazon_address   = AmazonAddress
  mtu              = MTU
  bgp_auth_key     = BGPAuthKey
  customer_address = CustomerAddress
  dx_gateway_id    = DirectConnectGatewayId
  tags             = TagList
  vpn_gateway_id   = AssociatedGatewayId
}

mapping "aws_dx_public_virtual_interface" {
  address_family        = AddressFamily
  bgp_asn               = ASN
  connection_id         = ConnectionId
  name                  = VirtualInterfaceName
  vlan                  = VLAN
  amazon_address        = AmazonAddress
  bgp_auth_key          = BGPAuthKey
  customer_address      = CustomerAddress
  route_filter_prefixes = RouteFilterPrefixList
  tags                  = TagList
}

test "aws_dx_bgp_peer" "address_family" {
  ok = "ipv4"
  ng = "ipv2"
}
