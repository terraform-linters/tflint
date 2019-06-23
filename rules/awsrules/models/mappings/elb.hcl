import = "aws-sdk-go/models/apis/elasticloadbalancing/2012-06-01/api-2.json"

mapping "aws_app_cookie_stickiness_policy" {
  name          = PolicyName
  load_balancer = AccessPointName
  lb_port       = AccessPointPort
  cookie_name   = CookieName
}

mapping "aws_elb" {
  name                        = AccessPointName
  name_prefix                 = any
  access_logs                 = AccessLog
  availability_zones          = AvailabilityZones
  security_groups             = SecurityGroups
  subnets                     = Subnets
  instances                   = Instances
  internal                    = any
  listener                    = Listeners
  health_check                = HealthCheck
  cross_zone_load_balancing   = CrossZoneLoadBalancingEnabled
  idle_timeout                = IdleTimeout
  connection_draining         = ConnectionDrainingEnabled
  connection_draining_timeout = ConnectionDrainingTimeout
  tags                        = TagList
}

mapping "aws_elb_attachment" {
  elb      = AccessPointName
  instance = InstanceId
}

mapping "aws_lb_cookie_stickiness_policy" {
  name                     = PolicyName
  load_balancer            = AccessPointName
  lb_port                  = AccessPointPort
  cookie_expiration_period = CookieExpirationPeriod
}

mapping "aws_lb_ssl_negotiation_policy" {
  name          = PolicyName
  load_balancer = AccessPointName
  lb_port       = AccessPointPort
  attribute     = PolicyAttributes
}

mapping "aws_load_balancer_backend_server_policy" {
  load_balancer_name = AccessPointName
  policy_names       = PolicyNames
  instance_port      = EndPointPort
}

mapping "aws_load_balancer_listener_policy" {
  load_balancer_name = AccessPointName
  load_balancer_port = AccessPointPort
  policy_names       = PolicyNames
}

mapping "aws_load_balancer_policy" {
  load_balancer_name = AccessPointName
  policy_name        = PolicyName
  policy_type_name   = PolicyTypeName
  policy_attribute   = PolicyAttributes
}

mapping "aws_proxy_protocol_policy" {
  load_balancer  = AccessPointName
  instance_ports = Ports
}
