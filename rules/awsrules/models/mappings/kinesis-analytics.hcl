import = "aws-sdk-go/models/apis/kinesisanalytics/2015-08-14/api-2.json"

mapping "aws_kinesis_analytics_application" {
  name                       = ApplicationName
  code                       = ApplicationCode
  description                = ApplicationDescription
  cloudwatch_logging_options = CloudWatchLoggingOptions
  inputs                     = Inputs
  outputs                    = Outputs
  reference_data_sources     = ReferenceDataSource
  tags                       = Tags
}
