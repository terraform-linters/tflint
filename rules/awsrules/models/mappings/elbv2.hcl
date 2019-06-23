import = "aws-sdk-go/models/apis/elasticloadbalancingv2/2015-12-01/api-2.json"

mapping "aws_lb" {
  name                             = LoadBalancerName
  name_prefix                      = any
  internal                         = any
  load_balancer_type               = LoadBalancerTypeEnum
  security_groups                  = SecurityGroups
  access_logs                      = any
  subnets                          = Subnets
  subnet_mapping                   = SubnetMappings
  idle_timeout                     = any
  enable_deletion_protection       = any
  enable_cross_zone_load_balancing = any
  enable_http2                     = any
  ip_address_type                  = IpAddressType
  tags                             = TagList
}

mapping "aws_alb" {
  name                             = LoadBalancerName
  name_prefix                      = any
  internal                         = any
  load_balancer_type               = LoadBalancerTypeEnum
  security_groups                  = SecurityGroups
  access_logs                      = any
  subnets                          = Subnets
  subnet_mapping                   = SubnetMappings
  idle_timeout                     = any
  enable_deletion_protection       = any
  enable_cross_zone_load_balancing = any
  enable_http2                     = any
  ip_address_type                  = IpAddressType
  tags                             = TagList
}

mapping "aws_lb_listener" {
  load_balancer_arn = LoadBalancerArn
  port              = Port
  protocol          = ProtocolEnum
  ssl_policy        = SslPolicyName
  certificate_arn   = CertificateList
  default_action    = Actions
}

mapping "aws_alb_listener" {
  load_balancer_arn = LoadBalancerArn
  port              = Port
  protocol          = ProtocolEnum
  ssl_policy        = SslPolicyName
  certificate_arn   = CertificateList
  default_action    = Actions
}

mapping "aws_lb_listener_certificate" {
  listener_arn    = ListenerArn
  certificate_arn = CertificateArn
}

mapping "aws_alb_listener_certificate" {
  listener_arn    = ListenerArn
  certificate_arn = CertificateArn
}

mapping "aws_lb_listener_rule" {
  listener_arn = ListenerArn
  priority     = RulePriority
  action       = Actions
  condition    = RuleConditionList
}

mapping "aws_alb_listener_rule" {
  listener_arn = ListenerArn
  priority     = RulePriority
  action       = Actions
  condition    = RuleConditionList
}

mapping "aws_lb_target_group" {
  name                               = TargetGroupName
  name_prefix                        = TargetGroupName
  port                               = Port
  protocol                           = ProtocolEnum
  vpc_id                             = VpcId
  deregistration_delay               = any
  slow_start                         = any
  lambda_multi_value_headers_enabled = any
  proxy_protocol_v2                  = any
  stickiness                         = any
  health_check                       = any
  target_type                        = TargetTypeEnum
  tags                               = TagList
}

mapping "aws_alb_target_group" {
  name                               = TargetGroupName
  name_prefix                        = TargetGroupName
  port                               = Port
  protocol                           = ProtocolEnum
  vpc_id                             = VpcId
  deregistration_delay               = any
  slow_start                         = any
  lambda_multi_value_headers_enabled = any
  proxy_protocol_v2                  = any
  stickiness                         = any
  health_check                       = any
  target_type                        = TargetTypeEnum
  tags                               = TagList
}

mapping "aws_lb_target_group_attachment" {
  target_group_arn  = TargetGroupArn
  target_id         = TargetId
  port              = Port
  availability_zone = ZoneName
}

mapping "aws_alb_target_group_attachment" {
  target_group_arn  = TargetGroupArn
  target_id         = TargetId
  port              = Port
  availability_zone = ZoneName
}

test "aws_lb" "ip_address_type" {
  ok = "ipv4"
  ng = "ipv6"
}

test "aws_lb" "load_balancer_type" {
  ok = "application"
  ng = "classic"
}

test "aws_lb_listener" "protocol" {
  ok = "HTTPS"
  ng = "UDP"
}

test "aws_lb_target_group" "target_type" {
  ok = "lambda"
  ng = "container"
}
