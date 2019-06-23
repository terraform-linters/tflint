import = "aws-sdk-go/models/apis/waf-regional/2016-11-28/api-2.json"

mapping "aws_wafregional_byte_match_set" {
  name              = ResourceName
  byte_match_tuples = ByteMatchTuples
}

mapping "aws_wafregional_geo_match_set" {
  name                 = ResourceName
  geo_match_constraint = GeoMatchConstraint
}

mapping "aws_wafregional_ipset" {
  name              = ResourceName
  ip_set_descriptor = IPSetDescriptor
}

mapping "aws_wafregional_rate_based_rule" {
  metric_name = MetricName
  name        = ResourceName
  rate_key    = RateKey
  rate_limit  = RateLimit
  predicate   = Predicate
}

mapping "aws_wafregional_regex_match_set" {
  name              = ResourceName
  regex_match_tuple = RegexMatchTuples
}

mapping "aws_wafregional_regex_pattern_set" {
  name                  = ResourceName
  regex_pattern_strings = RegexPatternStrings
}

mapping "aws_wafregional_rule" {
  name        = ResourceName
  metric_name = MetricName
  predicate   = Predicate
}

mapping "aws_wafregional_rule_group" {
  name           = ResourceName
  metric_name    = MetricName
  activated_rule = ActivatedRules
}

mapping "aws_wafregional_size_constraint_set" {
  name             = ResourceName
  size_constraints = SizeConstraints
}

mapping "aws_wafregional_sql_injection_match_set" {
  name                      = ResourceName
  sql_injection_match_tuple = SqlInjectionMatchTuples
}

mapping "aws_wafregional_web_acl" {
  default_action        = WafAction
  metric_name           = MetricName
  name                  = ResourceName
  logging_configuration = LoggingConfiguration
  rule                  = ActivatedRule
}

mapping "aws_wafregional_web_acl_association" {
  web_acl_id   = ResourceId
  resource_arn = ResourceArn
}

mapping "aws_wafregional_xss_match_set" {
  name            = ResourceName
  xss_match_tuple = XssMatchTuple
}
