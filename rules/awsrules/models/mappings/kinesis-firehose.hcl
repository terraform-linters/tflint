import = "aws-sdk-go/models/apis/firehose/2015-08-04/api-2.json"

mapping "aws_kinesis_firehose_delivery_stream" {
  name                         = DeliveryStreamName
  tags                         = TagDeliveryStreamInputTagList
  kinesis_source_configuration = KinesisStreamSourceConfiguration
  destination                  = any
  s3_configuration             = S3DestinationConfiguration
  extended_s3_configuration    = ExtendedS3DestinationConfiguration
  redshift_configuration       = RedshiftDestinationConfiguration
  elasticsearch_configuration  = ElasticsearchDestinationConfiguration
  splunk_configuration         = SplunkDestinationConfiguration
  cloudwatch_logging_options   = CloudWatchLoggingOptions
  processing_configuration     = ProcessingConfiguration
  processors                   = ProcessorList
  parameters                   = ProcessorParameterList
}
