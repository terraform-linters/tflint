import = "aws-sdk-go/models/apis/waf/2015-08-24/api-2.json"

mapping "aws_waf_byte_match_set" {
  name              = ResourceName
  byte_match_tuples = ByteMatchTuples
}

mapping "aws_waf_geo_match_set" {
  name                 = ResourceName
  geo_match_constraint = GeoMatchConstraint
}

mapping "aws_waf_ipset" {
  name               = ResourceName
  ip_set_descriptors = IPSetDescriptors
}

mapping "aws_waf_rate_based_rule" {
  metric_name = MetricName
  name        = ResourceName
  rate_key    = RateKey
  rate_limit  = RateLimit
  predicates  = Predicates
}

mapping "aws_waf_regex_match_set" {
  name              = ResourceName
  regex_match_tuple = RegexMatchTuples
}

mapping "aws_waf_regex_pattern_set" {
  name                  = ResourceName
  regex_pattern_strings = RegexPatternStrings
}

mapping "aws_waf_rule" {
  metric_name = MetricName
  name        = ResourceName
  predicates  = Predicates
}

mapping "aws_waf_rule_group" {
  name           = ResourceName
  metric_name    = MetricName
  activated_rule = ActivatedRules
}

mapping "aws_waf_size_constraint_set" {
  name             = ResourceName
  size_constraints = SizeConstraints
}

mapping "aws_waf_sql_injection_match_set" {
  name                       = ResourceName
  sql_injection_match_tuples = SqlInjectionMatchTuples
}

mapping "aws_waf_web_acl" {
  default_action        = WafAction
  metric_name           = MetricName
  name                  = ResourceName
  rules                 = ActivatedRules
  logging_configuration = LoggingConfiguration
}

mapping "aws_waf_xss_match_set" {
  name             = ResourceName
  xss_match_tuples = XssMatchTuples
}
