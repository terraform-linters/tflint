rule "aws_route_invalid_egress_only_gateway" {
    resource      = "aws_route"
    attribute     = "egress_only_gateway_id"
    source_action = "DescribeEgressOnlyInternetGateways"
    template      = "\"%s\" is invalid egress only internet gateway ID."
}

rule "aws_route_invalid_gateway" {
    resource      = "aws_route"
    attribute     = "gateway_id"
    source_action = "DescribeInternetGateways"
    template      = "\"%s\" is invalid internet gateway ID."
}

rule "aws_route_invalid_instance" {
    resource      = "aws_route"
    attribute     = "instance_id"
    source_action = "DescribeInstances"
    template      = "\"%s\" is invalid instance ID."
}

rule "aws_route_invalid_nat_gateway" {
    resource      = "aws_route"
    attribute     = "nat_gateway_id"
    source_action = "DescribeNatGateways"
    template      = "\"%s\" is invalid NAT gateway ID."
}

rule "aws_route_invalid_network_interface" {
    resource      = "aws_route"
    attribute     = "network_interface_id"
    source_action = "DescribeNetworkInterfaces"
    template      = "\"%s\" is invalid network interface ID."
}

rule "aws_route_invalid_route_table" {
    resource      = "aws_route"
    attribute     = "route_table_id"
    source_action = "DescribeRouteTables"
    template      = "\"%s\" is invalid route table ID."
}

rule "aws_route_invalid_vpc_peering_connection" {
    resource      = "aws_route"
    attribute     = "vpc_peering_connection_id"
    source_action = "DescribeVpcPeeringConnections"
    template      = "\"%s\" is invalid VPC peering connection ID."
}
