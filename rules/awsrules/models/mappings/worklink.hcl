import = "aws-sdk-go/models/apis/worklink/2018-09-25/api-2.json"

mapping "aws_worklink_fleet" {
  name                           = FleetName
  audit_stream_arn               = AuditStreamArn
  device_ca_certificate          = Certificate
  identity_provider              = any
  display_name                   = DisplayName
  network                        = any
  optimize_for_end_user_location = Boolean
}

mapping "aws_worklink_website_certificate_authority_association" {
  fleet_arn    = FleetArn
  certificate  = Certificate
  display_name = DisplayName
}
