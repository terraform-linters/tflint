import = "aws-sdk-go/models/apis/kafka/2018-11-14/api-2.json"

mapping "aws_msk_cluster" {
  broker_node_group_info = BrokerNodeGroupInfo
  cluster_name           = __stringMin1Max64
  kafka_version          = __stringMin1Max128
  number_of_broker_nodes = __integerMin1Max15
  client_authentication  = ClientAuthentication
  configuration_info     = ConfigurationInfo
  encryption_info        = EncryptionInfo
  enhanced_monitoring    = EnhancedMonitoring
  tags                   = __mapOf__string
}

mapping "aws_msk_configuration" {
  server_properties = __blob
  kafka_versions    = __listOf__string
  name              = __string
  description       = __string
}
