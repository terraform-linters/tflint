import = "aws-sdk-go/models/apis/email/2010-12-01/api-2.json"

mapping "aws_ses_active_receipt_rule_set" {
  rule_set_name = ReceiptRuleSetName
}

mapping "aws_ses_domain_identity" {
  domain = Domain
}

mapping "aws_ses_domain_identity_verification" {
  domain = Domain
}

mapping "aws_ses_domain_dkim" {
  domain = Domain
}

mapping "aws_ses_domain_mail_from" {
  domain                 = Domain
  mail_from_domain       = MailFromDomainName
  behavior_on_mx_failure = BehaviorOnMXFailure
}

mapping "aws_ses_email_identity" {
  email = Address
}

mapping "aws_ses_receipt_filter" {
  name   = ReceiptFilterName
  cidr   = Cidr
  policy = ReceiptFilterPolicy
}

mapping "aws_ses_receipt_rule" {
  name              = ReceiptRuleName
  rule_set_name     = ReceiptRuleSetName
  after             = ReceiptRuleName
  enabled           = Enabled
  recipients        = RecipientsList
  scan_enabled      = Enabled
  tls_policy        = TlsPolicy
  add_header_action = AddHeaderAction
  bounce_action     = BounceAction
  lambda_action     = LambdaAction
  s3_action         = S3Action
  sns_action        = SNSAction
  stop_action       = StopAction
  workmail_action   = WorkmailAction
}

mapping "aws_ses_receipt_rule_set" {
  rule_set_name = ReceiptRuleSetName
}

mapping "aws_ses_configuration_set" {
  name = ConfigurationSetName
}

mapping "aws_ses_event_destination" {
  name                   = EventDestinationName
  configuration_set_name = ConfigurationSetName
  enabled                = Enabled
  matching_types         = EventTypes
  cloudwatch_destination = CloudWatchDestination
  kinesis_destination    = KinesisFirehoseDestination
  sns_destination        = SNSDestination
}

mapping "aws_ses_identity_notification_topic" {
  topic_arn                = AmazonResourceName
  notification_type        = NotificationType
  identity                 = Identity
  include_original_headers = any
}

mapping "aws_ses_identity_policy" {
  identity = Identity
  name     = PolicyName
  policy   = Policy
}

mapping "aws_ses_template" {
  name    = TemplateName
  html    = HtmlPart
  subject = SubjectPart
  text    = TextPart
}
