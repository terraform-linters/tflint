import = "aws-sdk-go/models/apis/secretsmanager/2017-10-17/api-2.json"

mapping "aws_secretsmanager_secret" {
  name                    = NameType
  name_prefix             = any
  description             = DescriptionType
  kms_key_id              = KmsKeyIdType
  policy                  = NonEmptyResourcePolicyType
  recovery_window_in_days = RecoveryWindowInDaysType
  rotation_lambda_arn     = RotationLambdaARNType
  rotation_rules          = RotationRulesType
  tags                    = TagListType
}

mapping "aws_secretsmanager_secret_version" {
  secret_id      = SecretIdType
  secret_string  = SecretStringType
  secret_binary  = SecretBinaryType
  version_stages = SecretVersionStagesType
}
