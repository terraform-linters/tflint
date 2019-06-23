import = "aws-sdk-go/models/apis/pinpoint/2016-12-01/api-2.json"

mapping "aws_pinpoint_app" {
  name          = __string
  name_prefix   = any
  campaign_hook = CampaignHook
  limits        = CampaignLimits
  quiet_time    = QuietTime
}

mapping "aws_pinpoint_adm_channel" {
  application_id = __string
  client_id      = __string
  client_secret  = __string
  enabled        = __boolean
}

mapping "aws_pinpoint_apns_channel" {
  application_id                = __string
  enabled                       = __boolean
  default_authentication_method = __string
}

mapping "aws_pinpoint_apns_sandbox_channel" {
  application_id                = __string
  enabled                       = __boolean
  default_authentication_method = __string
}

mapping "aws_pinpoint_apns_voip_channel" {
  application_id                = __string
  enabled                       = __boolean
  default_authentication_method = __string
}

mapping "aws_pinpoint_apns_voip_sandbox_channel" {
  application_id                = __string
  enabled                       = __boolean
  default_authentication_method = __string
}

mapping "aws_pinpoint_baidu_channel" {
  application_id = __string
  enabled        = __boolean
  api_key        = __string
  secret_key     = __string
}

mapping "aws_pinpoint_email_channel" {
  application_id = __string
  enabled        = __boolean
  from_address   = __string
  identity       = __string
  role_arn       = __string
}

mapping "aws_pinpoint_event_stream" {
  application_id         = __string
  destination_stream_arn = __string
  role_arn               = __string
}

mapping "aws_pinpoint_gcm_channel" {
  application_id = __string
  api_key        = __string
  enabled        = __boolean
}

mapping "aws_pinpoint_sms_channel" {
  application_id = __string
  enabled        = __boolean
  sender_id      = __string
  short_code     = __string
}
