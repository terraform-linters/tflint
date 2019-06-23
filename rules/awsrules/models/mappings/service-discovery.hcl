import = "aws-sdk-go/models/apis/servicediscovery/2017-03-14/api-2.json"

mapping "aws_service_discovery_http_namespace" {
  name        = NamespaceName
  description = ResourceDescription
}

mapping "aws_service_discovery_private_dns_namespace" {
  name        = NamespaceName
  vpc         = ResourceId
  description = ResourceDescription
}

mapping "aws_service_discovery_public_dns_namespace" {
  name        = NamespaceName
  description = ResourceDescription
}

mapping "aws_service_discovery_service" {
  name                       = any // ServiceName
  description                = ResourceDescription
  dns_config                 = DnsConfig
  health_check_config        = HealthCheckConfig
  health_check_custom_config = HealthCheckCustomConfig
}
