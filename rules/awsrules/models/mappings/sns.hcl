import = "aws-sdk-go/models/apis/sns/2010-03-31/api-2.json"

mapping "aws_sns_platform_application" {
  name                             = String
  platform                         = String
  platform_credential              = any
  event_delivery_failure_topic_arn = topicARN
  event_endpoint_created_topic_arn = topicARN
  event_endpoint_deleted_topic_arn = topicARN
  event_endpoint_updated_topic_arn = topicARN
  failure_feedback_role_arn        = any
  platform_principal               = any
  success_feedback_role_arn        = any
  success_feedback_sample_rate     = any
}

mapping "aws_sns_sms_preferences" {
  monthly_spend_limit                   = any
  delivery_status_iam_role_arn          = any
  delivery_status_success_sampling_rate = any
  default_sender_id                     = any
  default_sms_type                      = any
  usage_report_s3_bucket                = any
}

mapping "aws_sns_topic" {
  name                                     = topicName
  name_prefix                              = any
  display_name                             = any
  policy                                   = any
  delivery_policy                          = any
  application_success_feedback_role_arn    = any
  application_success_feedback_sample_rate = any
  application_failure_feedback_role_arn    = any
  http_success_feedback_role_arn           = any
  http_success_feedback_sample_rate        = any
  http_failure_feedback_role_arn           = any
  kms_master_key_id                        = any
  lambda_success_feedback_role_arn         = any
  lambda_success_feedback_sample_rate      = any
  lambda_failure_feedback_role_arn         = any
  sqs_success_feedback_role_arn            = any
  sqs_success_feedback_sample_rate         = any
  sqs_failure_feedback_role_arn            = any
  tags                                     = TagList
}

mapping "aws_sns_topic_policy" {
  arn    = any
  policy = any
}

mapping "aws_sns_topic_subscription" {
  topic_arn                       = topicARN
  protocol                        = protocol
  endpoint                        = endpoint
  endpoint_auto_confirms          = any
  confirmation_timeout_in_minutes = any
  raw_message_delivery            = any
  filter_policy                   = any
  delivery_policy                 = any
}
