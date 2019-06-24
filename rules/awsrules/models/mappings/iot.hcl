import = "aws-sdk-go/models/apis/iot/2015-05-28/api-2.json"

mapping "aws_iot_certificate" {
  active = SetAsActive
  csr    = CertificateSigningRequest
}

mapping "aws_iot_policy" {
  name   = PolicyName
  policy = PolicyDocument
}

mapping "aws_iot_policy_attachment" {
  policy = PolicyName
  target = CertificateArn
}

mapping "aws_iot_topic_rule" {
  name              = RuleName
  description       = Description
  enabled           = IsDisabled
  sql               = SQL
  sql_version       = AwsIotSqlVersion
  cloudwatch_alarm  = CloudwatchAlarmAction
  cloudwatch_metric = CloudwatchMetricAction
  dynamodb          = DynamoDBAction
  elasticsearch     = ElasticsearchAction
  firehose          = FirehoseAction
  kinesis           = KinesisAction
  lambda            = LambdaAction
  republish         = RepublishAction
  s3                = S3Action
  sns               = SnsAction
  sqs               = SqsAction
}

mapping "aws_iot_thing" {
  name            = ThingName
  attributes      = Attributes
  thing_type_name = ThingTypeName
}

mapping "aws_iot_thing_principal_attachment" {
  principal = CertificateArn
  thing     = ThingName
}

mapping "aws_iot_thing_type" {
  name                  = ThingTypeName
  // description           = any // ThingTypeDescription
  deprecated            = UndoDeprecate
  // searchable_attributes = SearchableAttributes
}

mapping "aws_iot_role_alias" {
  alias               = RoleAlias
  role_arn            = RoleArn
  credential_duration = CredentialDurationSeconds
}
