import = "aws-sdk-go/models/apis/sqs/2012-11-05/api-2.json"

mapping "aws_sqs_queue" {
  name                              = String
  name_prefix                       = any
  visibility_timeout_seconds        = Integer
  message_retention_seconds         = Integer
  max_message_size                  = Integer
  delay_seconds                     = Integer
  receive_wait_time_seconds         = Integer
  policy                            = String
  redrive_policy                    = String
  fifo_queue                        = Boolean
  content_based_deduplication       = Boolean
  kms_master_key_id                 = String
  kms_data_key_reuse_period_seconds = Integer
  tags                              = TagMap
}

mapping "aws_sqs_queue_policy" {
  queue_url = String
  policy    = String
}
