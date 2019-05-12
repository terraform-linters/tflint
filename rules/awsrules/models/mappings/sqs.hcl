import = "aws-sdk-go/models/apis/sqs/2012-11-05/api-2.json"

mapping "aws_sqs_queue" {
  name                              = String
  name_prefix                       = any
  visibility_timeout_seconds        = String
  message_retention_seconds         = String
  max_message_size                  = String
  delay_seconds                     = String
  receive_wait_time_seconds         = String
  policy                            = String
  redrive_policy                    = String
  fifo_queue                        = String
  content_based_deduplication       = String
  kms_master_key_id                 = String
  kms_data_key_reuse_period_seconds = String
  tags                              = TagMap
}

mapping "aws_sqs_queue_policy" {
  queue_url = String
  policy    = String
}
