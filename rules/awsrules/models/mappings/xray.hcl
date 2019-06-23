import = "aws-sdk-go/models/apis/xray/2016-04-12/api-2.json"

mapping "aws_xray_sampling_rule" {
  rule_name      = RuleName
  resource_arn   = ResourceARN
  priority       = Priority
  fixed_rate     = FixedRate
  reservoir_size = ReservoirSize
  service_name   = ServiceName
  service_type   = ServiceType
  host           = Host
  http_method    = HTTPMethod
  url_path       = URLPath
  version        = Version
  attributes     = AttributeMap
}
