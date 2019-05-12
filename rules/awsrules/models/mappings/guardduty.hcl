import = "aws-sdk-go/models/apis/guardduty/2017-11-28/api-2.json"

mapping "aws_guardduty_detector" {
  enable                       = Boolean
  finding_publishing_frequency = FindingPublishingFrequency
}

mapping "aws_guardduty_invite_accepter" {
  detector_id       = DetectorId
  master_account_id = String
}

mapping "aws_guardduty_ipset" {
  activate    = Boolean
  detector_id = DetectorId
  format      = IpSetFormat
  location    = Location
  name        = Name
}

mapping "aws_guardduty_member" {
  account_id                 = String
  detector_id                = DetectorId
  email                      = Email
  invite                     = Boolean
  invitation_message         = String
  disable_email_notification = Boolean
}

mapping "aws_guardduty_threatintelset" {
  activate    = Boolean
  detector_id = DetectorId
  format      = ThreatIntelSetFormat
  location    = Location
  name        = Name
}
