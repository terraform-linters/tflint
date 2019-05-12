import = "aws-sdk-go/models/apis/mq/2017-11-27/api-2.json"

mapping "aws_mq_broker" {
  apply_immediately             = any
  auto_minor_version_upgrade    = __boolean
  broker_name                   = __string
  configuration                 = ConfigurationId
  deployment_mode               = DeploymentMode
  engine_type                   = EngineType
  engine_version                = __string
  host_instance_type            = __string
  publicly_accessible           = __boolean
  security_groups               = __listOf__string
  subnet_ids                    = __listOf__string
  maintenance_window_start_time = WeeklyStartTime
  logs                          = Logs
  user                          = __listOfUser
  tags                          = __mapOf__string
}

mapping "aws_mq_configuration" {
  data           = __string
  description    = __string
  engine_type    = EngineType
  engine_version = __string
  name           = __string
  tags           = __mapOf__string
}
