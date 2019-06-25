import = "aws-sdk-go/models/apis/kms/2014-11-01/api-2.json"

mapping "aws_kms_alias" {
  name          = AliasNameType
  name_prefix   = any
  target_key_id = KeyIdType
}

mapping "aws_kms_external_key" {
  deletion_window_in_days = any
  description             = DescriptionType
  enabled                 = BooleanType
  key_material_base64     = any
  policy                  = PolicyType
  tags                    = TagList
  valid_to                = DateType
}

mapping "aws_kms_grant" {
  name                  = GrantNameType
  key_id                = KeyIdType
  grantee_principal     = PrincipalIdType
  operations            = GrantOperationList
  retiring_principal    = PrincipalIdType
  constraints           = GrantConstraints
  grant_creation_tokens = GrantTokenList
  retire_on_delete      = BooleanType
}

mapping "aws_kms_ciphertext" {
  plaintext = PlaintextType
  key_id    = KeyIdType
  context   = EncryptionContextType
}

mapping "aws_kms_key" {
  description             = DescriptionType
  key_usage               = KeyUsageType
  policy                  = PolicyType
  deletion_window_in_days = any
  is_enabled              = BooleanType
  enable_key_rotation     = BooleanType
  tags                    = TagList
}
