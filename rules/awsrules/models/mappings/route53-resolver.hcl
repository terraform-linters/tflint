import = "aws-sdk-go/models/apis/route53resolver/2018-04-01/api-2.json"

mapping "aws_route53_resolver_endpoint" {
  direction          = ResolverEndpointDirection
  ip_address         = IpAddressesRequest
  security_group_ids = SecurityGroupIds
  name               = any // Name
  tags               = TagList
}

mapping "aws_route53_resolver_rule" {
  domain_name          = DomainName
  rule_type            = RuleTypeOption
  name                 = any // Name
  resolver_endpoint_id = ResourceId
  target_ip            = TargetList
  tags                 = TagList
}

mapping "aws_route53_resolver_rule_association" {
  resolver_rule_id = ResourceId
  vpc_id           = ResourceId
  name             = any // Name
}
